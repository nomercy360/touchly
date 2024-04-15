package transport

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"touchly/internal/db"
)

func getIDFromRequest(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func getUserIDFromRequest(r *http.Request) int64 {
	ctx := r.Context()

	return ctx.Value("userID").(int64)
}

// CreateContactHandler godoc
// @Summary      Create contact
// @Description  create contact
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Param        contact   body     db.Contact     true  "contact"
// @Success      201  {object}   db.Contact
// @Security     JWT
// @Router       /api/contacts [post]
func (tr *transport) CreateContactHandler(w http.ResponseWriter, r *http.Request) {
	var contact db.Contact
	if err := decodeRequest(r, &contact); err != nil {
		_ = WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := getUserIDFromRequest(r)

	createdContact, err := tr.api.CreateContact(userID, contact)
	if err != nil {
		_ = WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = WriteJSON(w, http.StatusCreated, createdContact)
}

// GetContactHandler godoc
// @Summary      Get contact
// @Description  get contact
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Param        id   path     int     true  "contact id"
// @Success      200  {object}   db.Contact
// @Router       /api/contacts/{id} [get]
func (tr *transport) GetContactHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		_ = WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	contact, err := tr.api.GetContact(id)

	if err != nil {
		if errors.As(err, &db.ErrNotFound) {
			_ = WriteError(w, http.StatusNotFound, err.Error())
			return
		}

		_ = WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = WriteJSON(w, http.StatusOK, contact)
}

// UpdateContactHandler godoc
// @Summary      Update contact
// @Description  update contact
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Param        id   path     int     true  "contact id"
// @Success      200  {object}   nil
// @Security     JWT
// @Router       /api/contacts/{id} [put]
func (tr *transport) UpdateContactHandler(w http.ResponseWriter, r *http.Request) {
	var contact db.Contact
	if err := decodeRequest(r, &contact); err != nil {
		_ = WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := getUserIDFromRequest(r)

	err := tr.api.UpdateContact(userID, contact)
	if err != nil {
		_ = WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = WriteJSON(w, http.StatusOK, nil)
}

// DeleteContactHandler godoc
// @Summary      Delete contact
// @Description  delete contact
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Param        id   path     int     true  "contact id"
// @Success      200  {object}   nil
// @Security     JWT
// @Router       /api/contacts/{id} [delete]
func (tr *transport) DeleteContactHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		_ = WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := getUserIDFromRequest(r)

	err = tr.api.DeleteContact(userID, id)
	if err != nil {
		_ = WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = WriteJSON(w, http.StatusOK, nil)
}

// ListContactsHandler godoc
// @Summary      List contacts
// @Description  get contacts
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Success      200  {array}   db.Contact
// @Router       /api/contacts [get]
func (tr *transport) ListContactsHandler(w http.ResponseWriter, r *http.Request) {
	contacts, err := tr.api.ListContacts()

	if err != nil {
		_ = WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = WriteJSON(w, http.StatusOK, contacts)
}

// ListSavedContactsHandler godoc
// @Summary      List contacts saved by user
// @Description  get saved contacts
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Param        id   path     int     true  "user id"
// @Success      200  {array}   db.Contact
// @Security     JWT
// @Router       /api/contacts/{id}/saved [get]
func (tr *transport) ListSavedContactsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		_ = WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	contacts, err := tr.api.ListSavedContacts(userID)
	if err != nil {
		_ = WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = WriteJSON(w, http.StatusOK, contacts)
}

type SaveContactRequest struct {
	ContactID int64 `json:"contact_id" example:"1"`
}

// SaveContactHandler godoc
// @Summary      Save contact
// @Description  save contact
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Param        id   path     int     true  "user id"
// @Param		 account	   body	   transport.SaveContactRequest	true	"contact id to save"
// @Success      200  {object}   nil
// @Security     JWT
// @Router       /api/contacts/{id}/save [post]
func (tr *transport) SaveContactHandler(w http.ResponseWriter, r *http.Request) {
	var data SaveContactRequest

	userID := getUserIDFromRequest(r)

	if err := decodeRequest(r, &data); err != nil {
		_ = WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := tr.api.SaveContact(userID, data.ContactID)
	if err != nil {
		_ = WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = WriteJSON(w, http.StatusOK, nil)
}

type DeleteSavedContactRequest struct {
	ContactID int64 `json:"contact_id" example:"1"`
}

// DeleteSavedContactHandler godoc
// @Summary      Delete saved contact
// @Description  delete saved contact
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Param        id   path     int     true  "user id"
// @Param		 account	   body	   transport.DeleteSavedContactRequest	true	"contact id to delete"
// @Success      200  {object}   nil
// @Security     JWT
// @Router       /api/contacts/{id}/saved [delete]
func (tr *transport) DeleteSavedContactHandler(w http.ResponseWriter, r *http.Request) {
	var data DeleteSavedContactRequest

	userID := getUserIDFromRequest(r)

	if err := decodeRequest(r, &data); err != nil {
		_ = WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := tr.api.DeleteSavedContact(userID, data.ContactID)
	if err != nil {
		_ = WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = WriteJSON(w, http.StatusOK, nil)
}

func (tr *transport) ListAddressesHandler(w http.ResponseWriter, r *http.Request) {
	addresses, err := tr.api.ListAddresses()

	if err != nil {
		_ = WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = WriteJSON(w, http.StatusOK, addresses)
}
