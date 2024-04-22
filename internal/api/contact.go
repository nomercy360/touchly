package api

import (
	"touchly/internal/db"
	"touchly/internal/terrors"
)

func (api *api) CreateContact(userID int64, contact db.Contact) (*db.Contact, error) {
	if contact.Name == "" {
		return nil, terrors.InvalidRequest(nil, "name is required")
	}

	contact.UserID = userID

	res, err := api.storage.CreateContact(contact)

	if err != nil {
		return nil, terrors.InternalServerError(err, "failed to create contact")
	}

	return res, nil
}

func (api *api) DeleteContact(userID, id int64) error {
	if err := api.storage.DeleteContact(userID, id); err != nil {
		return terrors.InternalServerError(err, "failed to delete contact")
	}

	return nil
}

func (api *api) UpdateContact(userID int64, contact db.Contact) error {
	contact.UserID = userID

	if err := api.storage.UpdateContact(contact); err != nil {
		return terrors.InternalServerError(err, "failed to update contact")
	}

	return nil
}

func (api *api) ListContacts(tagIDs []int, search string, lat float64, lng float64, radius int, page, pageSize int) (db.ContactsPage, error) {
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

	contacts, err := api.storage.ListContacts(tagIDs, search, lat, lng, radius, page, pageSize)

	if err != nil {
		return contacts, terrors.InternalServerError(err, "failed to list contacts")
	}

	return contacts, nil
}

func (api *api) GetContact(id int64) (*db.Contact, error) {
	contact, err := api.storage.GetContact(id)

	if err != nil {
		return nil, terrors.InternalServerError(err, "failed to get contact")
	}

	return contact, nil
}

func (api *api) SaveContact(userID, contactID int64) error {
	if err := api.storage.SaveContact(userID, contactID); err != nil {
		return terrors.InternalServerError(err, "failed to save contact")
	}

	return nil
}

func (api *api) DeleteSavedContact(userID, contactID int64) error {
	if err := api.storage.DeleteSavedContact(userID, contactID); err != nil {
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

func (api *api) CreateContactAddress(userID int64, address db.Address) (*db.Address, error) {
	if address.ContactID == 0 {
		return nil, terrors.InvalidRequest(nil, "contact id is required")
	}

	contact, err := api.storage.GetContact(address.ContactID)

	if err != nil {
		return nil, terrors.InternalServerError(err, "failed to get contact")
	}

	if contact.UserID != userID {
		return nil, terrors.Forbidden(nil, "contact does not belong to user")
	}

	res, err := api.storage.CreateContactAddress(address)

	if err != nil {
		return nil, terrors.InternalServerError(err, "failed to create contact address")
	}

	return res, nil
}
