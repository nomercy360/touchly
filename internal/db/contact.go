package db

import (
	"database/sql/driver"
	"fmt"
	"github.com/lib/pq"
	"strconv"
	"strings"
	"time"
)

type ContactVisibility string

const (
	ContactVisibilityPublic     ContactVisibility = "public"
	ContactVisibilityPrivate    ContactVisibility = "private"
	ContactVisibilitySharedLink ContactVisibility = "shared_link"
)

func (v ContactVisibility) IsValid() bool {
	switch v {
	case ContactVisibilityPublic, ContactVisibilityPrivate, ContactVisibilitySharedLink:
		return true
	}

	return false
}

type Contact struct {
	ID               int64             `db:"id" json:"id"`
	Name             string            `db:"name" json:"name"`
	Avatar           *string           `db:"avatar" json:"avatar"`
	ActivityName     *string           `db:"activity_name" json:"activity_name"`
	Website          *string           `db:"website" json:"website"`
	CountryCode      *string           `db:"country_code" json:"country_code"`
	About            *string           `db:"about" json:"about"`
	ViewsAmount      int               `db:"views_amount" json:"views_amount"`
	SavesAmount      int               `db:"saves_amount" json:"saves_amount"`
	CreatedAt        time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time         `db:"updated_at" json:"updated_at"`
	Address          *Address          `db:"-" json:"address"`
	PhoneNumber      string            `db:"phone_number" json:"phone_number"`
	PhoneCallingCode string            `db:"phone_calling_code" json:"phone_calling_code"`
	Email            string            `db:"email" json:"email"`
	Tags             []Tag             `db:"-" json:"tags"`
	SocialLinks      []Link            `db:"-" json:"social_links"`
	DeletedAt        *time.Time        `db:"deleted_at" json:"deleted_at"`
	UserID           int64             `db:"user_id" json:"user_id"`
	Visibility       ContactVisibility `db:"visibility" json:"visibility"`
}

type ContactListEntry struct {
	ID           int64             `db:"id" json:"id"`
	Name         string            `db:"name" json:"name"`
	Avatar       string            `db:"avatar" json:"avatar"`
	ActivityName string            `db:"activity_name" json:"activity_name"`
	About        string            `db:"about" json:"about"`
	ViewsAmount  int               `db:"views_amount" json:"views_amount"`
	SavesAmount  int               `db:"saves_amount" json:"saves_amount"`
	UserID       int64             `db:"user_id" json:"user_id"`
	Visibility   ContactVisibility `db:"visibility" json:"visibility"`
}

type ContactsPage struct {
	Contacts   []ContactListEntry `json:"contacts"`
	TotalCount int                `json:"total_count"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
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

func (s *storage) ListContacts(
	userID int64, tagIDs []int, search string, lat float64,
	lng float64, radius int, page, pageSize int) (ContactsPage, error) {

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

	if userID != 0 {
		whereClauses = append(whereClauses, "(c.user_id = $"+strconv.Itoa(paramIndex)+" OR c.visibility = 'public')")
	} else {
		whereClauses = append(whereClauses, "c.visibility = 'public'")
	}

	if len(tagIDs) > 0 {
		whereClauses = append(whereClauses, "ct.tag_id = ANY($"+strconv.Itoa(paramIndex)+")")
		args = append(args, pq.Array(tagIDs))
		paramIndex++
	}

	if lat != 0 && lng != 0 {
		point := fmt.Sprintf("ST_SetSRID(ST_Point(%f, %f), 4326)", lng, lat)

		geoClause := fmt.Sprintf("ST_DWithin(a.location, %s, $%d)", point, paramIndex)
		whereClauses = append(whereClauses, geoClause)
		args = append(args, radius*1000) // Convert km to meters
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

	if lat != 0 && lng != 0 {
		countQuery += ` JOIN addresses a ON c.id = a.contact_id`
	}

	countQuery += where
	err := s.pg.QueryRow(countQuery, args...).Scan(&contactsPage.TotalCount)
	if err != nil {
		return contactsPage, fmt.Errorf("error fetching contacts count: %w", err)
	}

	selectQuery := `
        SELECT c.id, c.name, c.avatar, c.activity_name, c.about, c.views_amount, c.saves_amount, c.user_id, c.visibility
        FROM contacts c`

	if len(tagIDs) > 0 {
		selectQuery += ` JOIN contact_tags ct ON c.id = ct.contact_id`
	}

	if lat != 0 && lng != 0 {
		selectQuery += ` JOIN addresses a ON c.id = a.contact_id`
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

	contacts := make([]ContactListEntry, 0)

	for rows.Next() {
		var c ContactListEntry
		err = rows.Scan(
			&c.ID, &c.Name, &c.Avatar, &c.ActivityName, &c.About, &c.ViewsAmount, &c.SavesAmount, &c.UserID,
			&c.Visibility,
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
	var res Contact

	query := `
		INSERT INTO contacts
		    (name, avatar, activity_name, about, website, country_code, phone_number, phone_calling_code, email, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, name, avatar, activity_name, about, website, country_code, phone_number, phone_calling_code, email, user_id, created_at, updated_at, visibility, deleted_at
	`

	err = tx.QueryRow(query, contact.Name, contact.Avatar, contact.ActivityName, contact.About, contact.Website, contact.CountryCode, contact.PhoneNumber, contact.PhoneCallingCode, contact.Email, contact.UserID).Scan(
		&res.ID, &res.Name, &res.Avatar, &res.ActivityName, &res.About, &res.Website, &res.CountryCode, &res.PhoneNumber, &res.PhoneCallingCode, &res.Email, &res.UserID, &res.CreatedAt, &res.UpdatedAt, &res.Visibility, &res.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	if len(contact.Tags) > 0 {
		for _, tag := range contact.Tags {
			if _, err = tx.Exec("INSERT INTO contact_tags (contact_id, tag_id) VALUES ($1, $2)", res.ID, tag.ID); err != nil {
				return nil, err
			}

			res.Tags = append(res.Tags, tag)
		}
	}

	if len(contact.SocialLinks) > 0 {
		for _, link := range contact.SocialLinks {
			row := tx.QueryRow("INSERT INTO social_media_links (contact_id, type, link) VALUES ($1, $2, $3) RETURNING id", res.ID, link.Type, link.Link)

			if err = row.Scan(&link.ID); err != nil {
				return nil, err
			}

			res.SocialLinks = append(res.SocialLinks, link)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &res, nil
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

func (s *storage) UpdateContact(userID, contactID int64, tags *[]Tag, links *[]Link, updates map[string]interface{}) (*Contact, error) {
	tx, err := s.pg.Beginx()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	baseQuery := "UPDATE contacts SET "
	var setClauses = []string{"updated_at = now()"}
	var queryParams = map[string]any{
		"id":      contactID,
		"user_id": userID,
	}

	for key, value := range updates {
		paramName := fmt.Sprintf("%s", key)
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", key, paramName))
		queryParams[paramName] = value
	}

	query := baseQuery + strings.Join(setClauses, ", ") + " WHERE id = :id AND user_id = :user_id RETURNING *"

	var contact Contact

	rows, err := s.pg.NamedQuery(query, queryParams)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		err = rows.StructScan(&contact)
		if err != nil {
			return nil, err
		}
	}

	if tags != nil {
		_, err = tx.Exec("DELETE FROM contact_tags WHERE contact_id=$1", contactID)
		if err != nil {
			return nil, err
		}

		for _, tag := range *tags {
			_, err = tx.Exec("INSERT INTO contact_tags (contact_id, tag_id) VALUES ($1, $2)", contactID, tag.ID)
			if err != nil {
				return nil, err
			}
		}
	}

	if links != nil {
		_, err = tx.Exec("DELETE FROM social_media_links WHERE contact_id=$1", contactID)
		if err != nil {
			return nil, err
		}

		for _, link := range *links {
			_, err = tx.Exec("INSERT INTO social_media_links (contact_id, type, link) VALUES ($1, $2, $3)", contactID, link.Type, link.Link)
			if err != nil {
				return nil, err
			}
		}
	}

	tagsUpdated := make([]Tag, 0)
	err = tx.Select(&tagsUpdated, "SELECT t.id, t.name FROM tags t JOIN contact_tags ct ON t.id = ct.tag_id WHERE ct.contact_id=$1", contactID)

	if err != nil {
		return nil, err
	}

	contact.Tags = tagsUpdated

	linksUpdated := make([]Link, 0)
	err = tx.Select(&linksUpdated, "SELECT id, type, link FROM social_media_links WHERE contact_id=$1", contactID)

	if err != nil {
		return nil, err
	}

	contact.SocialLinks = linksUpdated

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &contact, nil
}

func (s *storage) GetContact(userID, id int64) (*Contact, error) {
	var contact Contact

	query := `
		SELECT c.id, c.name, c.avatar, c.activity_name, c.about, c.views_amount,
		       c.saves_amount, c.created_at, c.updated_at, c.phone_number, c.email,
		       c.user_id, c.visibility, c.country_code, c.phone_calling_code, c.website, c.deleted_at
		FROM contacts c
		WHERE (c.id=$1 AND c.user_id=$2) OR (c.id=$1 AND c.visibility='public')
	`

	err := s.pg.Get(&contact, query, id, userID)

	if err != nil {
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
		if IsNoRowsError(err) {
			return &contact, nil
		}

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

func (s *storage) CreateContactAddress(address Address) (*Address, error) {
	query := `
		INSERT INTO addresses
			(external_id, contact_id, label, name, location)
		VALUES ($1, $2, $3, $4, ST_SetSRID(ST_Point($5, $6), 4326))
		RETURNING id
	`

	err := s.pg.QueryRow(query, address.ExternalID, address.ContactID, address.Label, address.Name, address.Location.Lng, address.Location.Lat).Scan(&address.ID)

	if err != nil {
		return nil, err
	}

	return &address, nil
}

func (s *storage) UpdateContactVisibility(userID, contactID int64, visibility ContactVisibility) error {
	_, err := s.pg.Exec("UPDATE contacts SET visibility=$1 WHERE id=$2 AND user_id=$3", visibility, contactID, userID)

	if err != nil {
		if IsNoRowsError(err) {
			return fmt.Errorf("not found")
		}

		return err
	}

	return nil
}

func (s *storage) GetContactsByUserID(userID int64) (ContactsPage, error) {
	contactsPage := ContactsPage{}

	query := `
		SELECT c.id, c.name, c.avatar, c.activity_name, c.about, c.views_amount, c.saves_amount, c.user_id, c.visibility
		FROM contacts c
		WHERE c.user_id=$1
	`

	rows, err := s.pg.Queryx(query, userID)

	if err != nil {
		return contactsPage, err
	}

	defer rows.Close()

	contacts := make([]ContactListEntry, 0)

	for rows.Next() {
		var c ContactListEntry
		err = rows.StructScan(&c)
		if err != nil {
			return contactsPage, err
		}
		contacts = append(contacts, c)
	}

	contactsPage.Contacts = contacts

	return contactsPage, nil
}
