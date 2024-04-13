package api

import (
	"errors"
	"touchly/internal/db"
)

func (api *api) CreateContact(userID int64, contact db.Contact) (*db.Contact, error) {
	if contact.Name == "" {
		return nil, errors.New("full name or name is required")
	}

	contact.UserID = userID

	res, err := api.storage.CreateContact(contact)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (api *api) DeleteContact(userID, id int64) error {
	if err := api.storage.DeleteContact(userID, id); err != nil {
		return err
	}

	return nil
}

func (api *api) UpdateContact(userID int64, contact db.Contact) error {
	contact.UserID = userID

	if err := api.storage.UpdateContact(contact); err != nil {
		return err
	}

	return nil
}

func (api *api) ListContacts() ([]db.Contact, error) {
	contacts, err := api.storage.ListContacts()

	if err != nil {
		return nil, err
	}

	return contacts, nil
}

func (api *api) GetContact(id int64) (*db.Contact, error) {
	contact, err := api.storage.GetContact(id)

	if err != nil {
		return nil, err
	}

	return contact, nil
}

func (api *api) SaveContact(userID, contactID int64) error {
	if err := api.storage.SaveContact(userID, contactID); err != nil {
		return err
	}

	return nil
}

func (api *api) DeleteSavedContact(userID, contactID int64) error {
	if err := api.storage.DeleteSavedContact(userID, contactID); err != nil {
		return err
	}

	return nil
}

func (api *api) ListSavedContacts(userID int64) ([]db.Contact, error) {
	contacts, err := api.storage.ListSavedContacts(userID)

	if err != nil {
		return nil, err
	}

	return contacts, nil
}

func (api *api) ListAddresses() ([]db.Address, error) {
	addresses, err := api.storage.ListAddresses()

	if err != nil {
		return nil, err
	}

	return addresses, nil
}
