package transport

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"strings"
	api2 "touchly/internal/api"
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
		WriteError(r, w, err)
		return
	}

	userID := getUserIDFromRequest(r)

	createdContact, err := tr.api.CreateContact(userID, contact)
	if err != nil {
		WriteError(r, w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, createdContact)
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
		WriteError(r, w, err)
		return
	}

	contact, err := tr.api.GetContact(id)

	if err != nil {
		WriteError(r, w, err)
		return
	}

	WriteJSON(w, http.StatusOK, contact)
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
	var contact api2.UpdateContactRequest
	if err := decodeRequest(r, &contact); err != nil {
		WriteError(r, w, err)
		return
	}

	userID := getUserIDFromRequest(r)
	cID, _ := getIDFromRequest(r)

	res, err := tr.api.UpdateContact(userID, cID, contact)

	if err != nil {
		WriteError(r, w, err)
		return
	}

	WriteJSON(w, http.StatusOK, res)
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
		WriteError(r, w, err)
		return
	}

	userID := getUserIDFromRequest(r)

	err = tr.api.DeleteContact(userID, id)
	if err != nil {
		WriteError(r, w, err)
		return
	}

	WriteOK(w)
}

func queryToIntArray(query string) ([]int, error) {
	if query == "" {
		return nil, nil
	}

	queryArray := strings.Split(query, ",")

	var intArray []int

	for _, q := range queryArray {
		i, err := strconv.Atoi(q)
		if err != nil {
			return nil, err
		}

		intArray = append(intArray, i)
	}

	return intArray, nil
}

// ListContactsHandler godoc
// @Summary      List contacts
// @Description  get contacts
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Success      200  {object} db.ContactsPage
// @Param        page      query    int     false  "page number (default 1)"
// @Param        page_size query    int     false  "page size (default 20)"
// @Param		 search    query    string  false  "search query, search by name or activity"
// @Param		 tag       query    []int   false  "tag id"
// @Param 		 lat       query    float64 false  "latitude"
// @Param		 lng       query    float64 false  "longitude"
// @Param        radius    query    int     false  "radius in km"
// @Router       /api/contacts [get]
func (tr *transport) ListContactsHandler(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	search := r.URL.Query().Get("search")

	tags := r.URL.Query().Get("tag")

	tagIDs, _ := queryToIntArray(tags)

	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)

	radius, _ := strconv.Atoi(r.URL.Query().Get("radius"))

	contacts, err := tr.api.ListContacts(tagIDs, search, lat, lng, radius, page, pageSize)

	if err != nil {
		WriteError(r, w, err)
		return
	}

	WriteJSON(w, http.StatusOK, contacts)
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
// @Router       /api/contacts/saved [get]
func (tr *transport) ListSavedContactsHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromRequest(r)

	contacts, err := tr.api.ListSavedContacts(userID)

	if err != nil {
		WriteError(r, w, err)
		return
	}

	WriteJSON(w, http.StatusOK, contacts)
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
		WriteError(r, w, err)
		return
	}

	err := tr.api.SaveContact(userID, data.ContactID)
	if err != nil {
		WriteError(r, w, err)
		return
	}

	WriteOK(w)
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
		WriteError(r, w, err)
		return
	}

	err := tr.api.DeleteSavedContact(userID, data.ContactID)
	if err != nil {
		WriteError(r, w, err)
		return
	}

	WriteOK(w)
}

// CreateContactAddressHandler godoc
// @Summary      Create contact address
// @Description  create contact address
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Param        id		path     int     true  "contact id"
// @Param        address   body     db.Address     true  "address"
// @Success      201  {object}   db.Address
// @Security     JWT
// @Router       /api/contacts/{id}/addresses [post]
func (tr *transport) CreateContactAddressHandler(w http.ResponseWriter, r *http.Request) {
	var address db.Address
	if err := decodeRequest(r, &address); err != nil {
		WriteError(r, w, err)
		return
	}

	userID := getUserIDFromRequest(r)
	contactID, _ := getIDFromRequest(r)

	address.ContactID = contactID

	createdAddress, err := tr.api.CreateContactAddress(userID, address)
	if err != nil {
		WriteError(r, w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, createdAddress)
}
