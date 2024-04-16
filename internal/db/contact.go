package db

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"strconv"
	"strings"
	"time"
)

type Contact struct {
	ID               int64      `db:"id" json:"id"`
	Name             string     `db:"name" json:"name"`
	Avatar           *string    `db:"avatar" json:"avatar"`
	ActivityName     *string    `db:"activity_name" json:"activity_name"`
	About            *string    `db:"about" json:"about"`
	ViewsAmount      int        `db:"views_amount" json:"views_amount"`
	SavesAmount      int        `db:"saves_amount" json:"saves_amount"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
	Address          *Address   `db:"-" json:"address,omitempty"`
	PhoneNumber      string     `db:"phone_number" json:"phone_number"`
	PhoneCallingCode string     `db:"phone_calling_code" json:"phone_calling_code"`
	Email            string     `db:"email" json:"email"`
	Tags             []Tag      `db:"-" json:"tags,omitempty"`
	SocialLinks      []Link     `db:"-" json:"social_links,omitempty"`
	DeletedAt        *time.Time `db:"deleted_at" json:"deleted_at"`
	UserID           int64      `db:"user_id" json:"user_id"`
}

type ContactsPage struct {
	Contacts   []Contact `json:"contacts"`
	TotalCount int       `json:"total_count"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
}

type Link struct {
	ID        int64  `db:"id" json:"id"`
	Type      string `db:"type" json:"type"`
	Link      string `db:"link" json:"link"`
	ContactID int64  `db:"contact_id" json:"contact_id"`
	Label     string `db:"label" json:"label"`
}

type Tag struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (l *Point) Value() (driver.Value, error) {
	return fmt.Sprintf("POINT(%f %f)", l.Lng, l.Lat), nil
}

func (l *Point) Scan(src interface{}) error {
	if src == nil {
		return nil
	}

	switch src := src.(type) {
	case string:
		// Parse POINT(-73.935242 40.730610)
		return parsePoint(src, l)
	case []byte:
		// Same parsing, but converts []byte to string first
		return parsePoint(string(src), l)
	}

	return fmt.Errorf("cannot scan type %T into Location: %v", src, src)
}

func parsePoint(src string, p *Point) error {
	src = strings.TrimPrefix(src, "POINT(")
	src = strings.TrimSuffix(src, ")")
	parts := strings.Split(src, " ")
	if len(parts) != 2 {
		return fmt.Errorf("invalid POINT data")
	}

	var err error
	if p.Lng, err = strconv.ParseFloat(parts[0], 64); err != nil {
		return err
	}

	if p.Lat, err = strconv.ParseFloat(parts[1], 64); err != nil {
		return err
	}

	return nil
}

type Address struct {
	ID         int64      `db:"id" json:"id"`
	ExternalID string     `db:"external_id" json:"external_id"`
	ContactID  int64      `db:"contact_id" json:"contact_id"`
	Label      string     `db:"label" json:"label"`
	Name       string     `db:"name" json:"name"`
	Location   Point      `db:"location" json:"location"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at" json:"deleted_at"`
}

func (s *storage) ListContacts(tagIDs []int, search string, page, pageSize int) (ContactsPage, error) {
	contactsPage := ContactsPage{
		Page:     page,
		PageSize: pageSize,
	}

	offset := (page - 1) * pageSize

	var args []interface{}
	paramIndex := 1

	var whereClauses []string
	if search != "" {
		whereClauses = append(whereClauses, "(c.name ILIKE $"+strconv.Itoa(paramIndex)+" OR c.activity_name ILIKE $"+strconv.Itoa(paramIndex)+")")
		args = append(args, "%"+search+"%")
		paramIndex++
	}

	if len(tagIDs) > 0 {
		whereClauses = append(whereClauses, "ct.tag_id = ANY($"+strconv.Itoa(paramIndex)+")")
		args = append(args, pq.Array(tagIDs))
		paramIndex++
	}

	where := ""
	if len(whereClauses) > 0 {
		where = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	countQuery := `SELECT COUNT(*) FROM contacts c`
	if len(tagIDs) > 0 {
		countQuery += ` JOIN contact_tags ct ON c.id = ct.contact_id`
	}

	countQuery += where
	err := s.pg.QueryRow(countQuery, args...).Scan(&contactsPage.TotalCount)
	if err != nil {
		return contactsPage, fmt.Errorf("error fetching contacts count: %w", err)
	}

	selectQuery := `
        SELECT c.id, c.name, c.avatar, c.activity_name, c.about, c.views_amount, c.saves_amount, c.created_at, c.updated_at, c.phone_number, c.email, c.user_id
        FROM contacts c`

	if len(tagIDs) > 0 {
		selectQuery += ` JOIN contact_tags ct ON c.id = ct.contact_id`
	}

	selectQuery += where
	selectQuery += `
        ORDER BY c.created_at DESC
        LIMIT $` + strconv.Itoa(paramIndex) + ` OFFSET $` + strconv.Itoa(paramIndex+1)

	args = append(args, pageSize, offset)

	rows, err := s.pg.Query(selectQuery, args...)
	if err != nil {
		return contactsPage, err
	}

	defer rows.Close()

	contacts := make([]Contact, 0)

	for rows.Next() {
		var c Contact
		err = rows.Scan(
			&c.ID, &c.Name, &c.Avatar, &c.ActivityName, &c.About, &c.ViewsAmount, &c.SavesAmount, &c.CreatedAt, &c.UpdatedAt, &c.PhoneNumber, &c.Email, &c.UserID,
		)
		if err != nil {
			return contactsPage, fmt.Errorf("scanning contact row: %w", err)
		}
		contacts = append(contacts, c)
	}

	contactsPage.Contacts = contacts
	return contactsPage, nil
}

func (s *storage) CreateContact(contact Contact) (*Contact, error) {
	tx, err := s.pg.Beginx()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	// Insert the contact
	var contactID int64
	query := `
		INSERT INTO contacts
		    (name, avatar, activity_name, about, website, country_code, deleted_at, phone_number, phone_calling_code, email, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`

	err = tx.QueryRow(query, contact.Name, contact.Avatar, contact.ActivityName, contact.About, contact.Address, contact.PhoneNumber, contact.Email).Scan(&contactID)

	if err != nil {
		return nil, err
	}

	if len(contact.Tags) > 0 {
		for _, tag := range contact.Tags {
			if _, err = tx.Exec("INSERT INTO contact_tags (contact_id, tag_id) VALUES ($1, $2)", contactID, tag.ID); err != nil {
				return nil, err
			}
		}
	}

	if len(contact.SocialLinks) > 0 {
		for _, link := range contact.SocialLinks {
			row := tx.QueryRow("INSERT INTO social_media_links (contact_id, type, link) VALUES ($1, $2, $3) RETURNING id", contactID, link.Type, link.Link)

			if err = row.Scan(&link.ID); err != nil {
				return nil, err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	contact.ID = contactID

	return &contact, nil
}

func (s *storage) DeleteContact(userID, id int64) error {
	res, err := s.pg.Exec("DELETE FROM contacts WHERE id=$1 AND user_id=$2", id, userID)

	if err != nil {
		return err
	}

	rowsAffected, _ := res.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("contact with id %d not found", id)
	}

	return nil
}

func (s *storage) UpdateContact(contact Contact) error {
	tx, err := s.pg.Beginx()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	query := `
		UPDATE contacts
		SET name=$1, avatar=$2, activity_name=$3, about=$4, website=$5, country_code=$6, phone_number=$7, phone_calling_code=$8, email=$9
		WHERE id=$10
	`

	_, err = tx.Exec(query, contact.Name, contact.Avatar, contact.ActivityName, contact.About, contact.Address, contact.PhoneNumber, contact.Email, contact.ID)

	if err != nil {
		return err
	}

	// Update tags
	if len(contact.Tags) > 0 {
		if _, err := tx.Exec("DELETE FROM contact_tags WHERE contact_id=$1", contact.ID); err != nil {
			return err
		}

		for _, tag := range contact.Tags {
			if _, err := tx.Exec("INSERT INTO contact_tags (contact_id, tag_id) VALUES ($1, $2)", contact.ID, tag.ID); err != nil {
				return err
			}
		}
	}

	// Update social links
	if len(contact.SocialLinks) > 0 {
		if _, err := tx.Exec("DELETE FROM social_media_links WHERE contact_id=$1", contact.ID); err != nil {
			return err
		}

		for _, link := range contact.SocialLinks {
			if _, err := tx.Exec("INSERT INTO social_media_links (contact_id, type, link) VALUES ($1, $2, $3)", contact.ID, link.Type, link.Link); err != nil {
				return err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *storage) GetContact(id int64) (*Contact, error) {
	var contact Contact

	query := `
		SELECT c.id, c.name, c.avatar, c.activity_name, c.about, c.views_amount, c.saves_amount, c.created_at, c.updated_at, c.phone_number, c.email, c.user_id
		FROM contacts c
		WHERE c.id=$1
	`

	err := s.pg.Get(&contact, query, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("contact with id %d not found", id)
		}
		return nil, err
	}

	tags := make([]Tag, 0)

	err = s.pg.Select(&tags, "SELECT t.id, t.name FROM tags t JOIN contact_tags ct ON t.id = ct.tag_id WHERE ct.contact_id=$1", id)

	if err != nil {
		return nil, err
	}

	contact.Tags = tags
	links := make([]Link, 0)

	err = s.pg.Select(&links, "SELECT id, type, link FROM social_media_links WHERE contact_id=$1", id)

	if err != nil {
		return nil, err
	}

	contact.SocialLinks = links
	var address Address

	err = s.pg.Get(&address, "SELECT id, external_id, contact_id, label, name, ST_AsText(location) as location, created_at, updated_at, deleted_at FROM addresses WHERE contact_id=$1", id)

	if err != nil {
		return nil, err
	}

	contact.Address = &address

	return &contact, nil
}

func (s *storage) SaveContact(userID, contactID int64) error {
	_, err := s.pg.Exec("INSERT INTO saved_contacts (user_id, contact_id) VALUES ($1, $2)", userID, contactID)

	if err != nil {
		return err
	}

	return nil
}

func (s *storage) DeleteSavedContact(userID, contactID int64) error {
	_, err := s.pg.Exec("DELETE FROM saved_contacts WHERE user_id=$1 AND contact_id=$2", userID, contactID)

	if err != nil {
		return err
	}

	return nil
}

func (s *storage) ListSavedContacts(userID int64) ([]Contact, error) {
	contacts := make([]Contact, 0)

	query := `
		SELECT c.id, c.name, c.avatar, c.activity_name, c.about, c.views_amount, c.saves_amount, c.created_at, c.updated_at, c.phone_number, c.email, c.user_id
		FROM contacts c
		WHERE c.id IN (SELECT contact_id FROM saved_contacts WHERE user_id=$1)
	`

	err := s.pg.Select(&contacts, query, userID)

	if err != nil {
		return nil, err
	}

	return contacts, nil
}
