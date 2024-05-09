package api

import (
	"errors"
	"fmt"
	"touchly/internal/db"
	"touchly/internal/terrors"
)

type CreateAddressRequest struct {
	ExternalID *string `json:"external_id"`
	Label      string  `json:"label" validate:"required"`
	Name       string  `json:"name" validate:"required"`
	Lat        float64 `json:"lat" validate:"required"`
	Lng        float64 `json:"lng" validate:"required"`
} // @Name CreateAddressRequest

func (a CreateAddressRequest) toAddress() db.Address {
	return db.Address{
		ExternalID: a.ExternalID,
		Label:      a.Label,
		Name:       a.Name,
		Location: db.Point{
			Lat: a.Lat,
			Lng: a.Lng,
		},
	}
}

func (api *api) CreateContact(userID int64, contact CreateContactRequest) (*db.Contact, error) {
	res, err := api.storage.CreateContact(userID, contact.toContact(), contact.Tags, contact.SocialLinks)

	if err != nil {
		return nil, terrors.InternalServerError(err, "failed to create contact")
	}

	return res, nil
}

func (api *api) DeleteContact(userID, id int64) error {
	err := api.storage.DeleteContact(userID, id)

	if err != nil && errors.Is(err, db.ErrNotFound) {
		return terrors.NotFound(err, "contact not found")
	} else if err != nil {
		return terrors.InternalServerError(err, "failed to delete contact")
	}

	return nil
}

type UpdateContactRequest struct {
	Name             *string    `db:"name" json:"name" validate:"required"`
	Avatar           *string    `db:"avatar" json:"avatar"`
	ActivityName     *string    `db:"activity_name" json:"activity_name"`
	Website          *string    `db:"website" json:"website"`
	CountryCode      *string    `db:"country_code" json:"country_code"`
	About            *string    `db:"about" json:"about"`
	PhoneNumber      *string    `db:"phone_number" json:"phone_number"`
	PhoneCallingCode *string    `db:"phone_calling_code" json:"phone_calling_code"`
	Email            *string    `db:"email" json:"email"`
	Tags             *[]db.Tag  `db:"-" json:"tags,omitempty"`
	SocialLinks      *[]db.Link `db:"-" json:"social_links,omitempty"`
} // @Name UpdateContactRequest

type CreateContactRequest UpdateContactRequest // @Name CreateContactRequest

func (r CreateContactRequest) toContact() db.Contact {
	return db.Contact{
		Name:             *r.Name,
		Avatar:           r.Avatar,
		ActivityName:     r.ActivityName,
		Website:          r.Website,
		CountryCode:      r.CountryCode,
		About:            r.About,
		PhoneNumber:      r.PhoneNumber,
		PhoneCallingCode: r.PhoneCallingCode,
		Email:            r.Email,
	}
}

func collectUpdates(contact UpdateContactRequest) map[string]interface{} {
	updates := map[string]interface{}{}

	if contact.Name != nil {
		updates["name"] = *contact.Name
	}

	if contact.Avatar != nil {
		updates["avatar"] = *contact.Avatar
	}

	if contact.ActivityName != nil {
		updates["activity_name"] = *contact.ActivityName
	}

	if contact.Website != nil {
		updates["website"] = *contact.Website
	}

	if contact.CountryCode != nil {
		updates["country_code"] = *contact.CountryCode
	}

	if contact.About != nil {
		updates["about"] = *contact.About
	}

	if contact.PhoneNumber != nil {
		updates["phone_number"] = *contact.PhoneNumber
	}

	if contact.PhoneCallingCode != nil {
		updates["phone_calling_code"] = *contact.PhoneCallingCode
	}

	if contact.Email != nil {
		updates["email"] = *contact.Email
	}

	return updates
}

func (api *api) UpdateContact(userID, contactID int64, request UpdateContactRequest) (*db.Contact, error) {
	updates := collectUpdates(request)

	res, err := api.storage.UpdateContact(userID, contactID, request.Tags, request.SocialLinks, updates)

	if err != nil {
		return nil, terrors.InternalServerError(err, "failed to update contact")
	}

	return res, nil
}

func (api *api) ListContacts(userID int64, tagIDs []int, search string, lat float64, lng float64, radius int, page, pageSize int) (db.ContactsPage, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 20
	}

	if lat != 0 && lng != 0 {
		if radius < 1 {
			return db.ContactsPage{}, terrors.InvalidRequest(nil, "radius is required")
		}
	}

	query := db.ContactQuery{
		TagIDs:   tagIDs,
		Search:   search,
		Lat:      lat,
		Lng:      lng,
		Radius:   radius,
		Page:     page,
		PageSize: pageSize,
		UserID:   userID,
	}

	contacts, err := api.storage.ListContacts(query)

	if err != nil {
		return contacts, terrors.InternalServerError(err, "failed to list contacts")
	}

	return contacts, nil
}

func (api *api) GetContact(userID, id int64) (*db.Contact, error) {
	contact, err := api.storage.GetContact(userID, id)

	if err != nil {
		if db.IsNoRowsError(err) {
			return nil, terrors.NotFound(fmt.Errorf("contact not found"), "contact not found")
		}

		return nil, terrors.InternalServerError(err, "failed to get contact")
	}

	return contact, nil
}

func (api *api) SaveContact(userID, contactID int64) error {
	err := api.storage.SaveContact(userID, contactID)

	if err != nil && errors.Is(err, db.ErrAlreadyExists) {
		return terrors.InvalidRequest(err, "contact already saved")
	} else if err != nil {
		return terrors.InternalServerError(err, "failed to save contact")
	}

	return nil
}

func (api *api) DeleteSavedContact(userID, contactID int64) error {
	err := api.storage.DeleteSavedContact(userID, contactID)

	if err != nil && errors.Is(err, db.ErrNotFound) {
		return terrors.NotFound(err, "contact not found")
	} else if err != nil {
		return terrors.InternalServerError(err, "failed to delete saved contact")
	}

	return nil
}

func (api *api) ListSavedContacts(userID int64) ([]db.Contact, error) {
	contacts, err := api.storage.ListSavedContacts(userID)

	if err != nil {
		return nil, terrors.InternalServerError(err, "failed to list saved contacts")
	}

	return contacts, nil
}

func (api *api) CreateContactAddress(userID, contactID int64, address CreateAddressRequest) (*db.Address, error) {
	contact, err := api.storage.GetContact(userID, contactID)

	if err != nil && db.IsNoRowsError(err) {
		return nil, terrors.NotFound(err, "contact not found")
	} else if err != nil {
		return nil, terrors.InternalServerError(err, "failed to get contact")
	}

	if contact.Address != nil {
		return nil, terrors.InvalidRequest(nil, "contact already has an address")
	}

	if contact.UserID != userID {
		return nil, terrors.Forbidden(nil, "contact does not belong to user")
	}

	res, err := api.storage.CreateContactAddress(contactID, address.toAddress())

	if err != nil {
		return nil, terrors.InternalServerError(err, "failed to create contact address")
	}

	return res, nil
}

func (api *api) UpdateContactVisibility(userID, contactID int64, visibility db.ContactVisibility) error {
	if !visibility.IsValid() {
		return terrors.InvalidRequest(nil, "invalid visibility value")
	}

	if err := api.storage.UpdateContactVisibility(userID, contactID, visibility); err != nil {
		return terrors.InternalServerError(err, "failed to update contact visibility")
	}

	return nil
}

func (api *api) ListMyContacts(userID int64) (db.ContactsPage, error) {
	contacts, err := api.storage.GetContactsByUserID(userID)

	if err != nil {
		return db.ContactsPage{}, terrors.InternalServerError(err, "failed to get contacts")
	}

	return contacts, nil
}
