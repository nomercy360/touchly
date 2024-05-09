package handler

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"strings"
	api2 "touchly/internal/api"
	"touchly/internal/db"
)

func getUserID(c echo.Context) int64 {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*api2.JWTClaims)
	return claims.UserID
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
func (tr *transport) CreateContactHandler(c echo.Context) error {
	var contact db.Contact
	if err := c.Bind(&contact); err != nil {
		return err
	}

	userID := getUserID(c)

	createdContact, err := tr.api.CreateContact(userID, contact)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, createdContact)
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
func (tr *transport) GetContactHandler(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := getUserID(c)

	contact, err := tr.api.GetContact(userID, id)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, contact)
}

// UpdateContactHandler godoc
// @Summary      Update contact
// @Description  update contact
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Param        id   path     int     true  "contact id"
// @Param        contact   body     UpdateContactRequest     true  "contact"
// @Success      200  {object}   nil
// @Security     JWT
// @Router       /api/contacts/{id} [put]
func (tr *transport) UpdateContactHandler(c echo.Context) error {
	var contact api2.UpdateContactRequest
	if err := c.Bind(&contact); err != nil {
		return err
	}

	userID := getUserID(c)
	cID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	res, err := tr.api.UpdateContact(userID, cID, contact)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
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
func (tr *transport) DeleteContactHandler(c echo.Context) error {
	id, _ := getID(c)

	userID := getUserID(c)

	if err := tr.api.DeleteContact(userID, id); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
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
func (tr *transport) ListContactsHandler(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))

	search := c.QueryParam("search")

	tags := c.QueryParam("tag")

	tagIDs, _ := queryToIntArray(tags)
	userID := getUserID(c)

	lat, _ := strconv.ParseFloat(c.QueryParam("lat"), 64)
	lng, _ := strconv.ParseFloat(c.QueryParam("lng"), 64)

	radius, _ := strconv.Atoi(c.QueryParam("radius"))

	contacts, err := tr.api.ListContacts(userID, tagIDs, search, lat, lng, radius, page, pageSize)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, contacts)
}

// ListSavedContactsHandler godoc
// @Summary      List contacts saved by user
// @Description  get saved contacts
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Success      200  {array}   db.Contact
// @Security     JWT
// @Router       /api/contacts/saved [get]
func (tr *transport) ListSavedContactsHandler(c echo.Context) error {
	userID := getUserID(c)

	contacts, err := tr.api.ListSavedContacts(userID)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, contacts)
}

// SaveContactHandler godoc
// @Summary      Save contact
// @Description  save contact
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Param        id   path     int     true  "contact id"
// @Success      200  {object}   nil
// @Security     JWT
// @Router       /api/contacts/{id}/save [post]
func (tr *transport) SaveContactHandler(c echo.Context) error {
	userID := getUserID(c)

	contID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	err := tr.api.SaveContact(userID, contID)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusCreated)
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
// @Param		 account	   body	   handler.DeleteSavedContactRequest	true	"contact id to delete"
// @Success      200  {object}   nil
// @Security     JWT
// @Router       /api/contacts/{id}/saved [delete]
func (tr *transport) DeleteSavedContactHandler(c echo.Context) error {
	var data DeleteSavedContactRequest

	userID := getUserID(c)

	if err := c.Bind(&data); err != nil {
		return err
	}

	err := tr.api.DeleteSavedContact(userID, data.ContactID)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
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
// @Router       /api/contacts/{id}/address [post]
func (tr *transport) CreateContactAddressHandler(c echo.Context) error {
	var address db.Address
	if err := c.Bind(&address); err != nil {
		return err
	}

	if err := c.Validate(address); err != nil {
		return err
	}

	userID := getUserID(c)
	contactID, _ := getID(c)

	createdAddress, err := tr.api.CreateContactAddress(userID, contactID, address)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, createdAddress)
}

type UpdateContactVisibilityRequest struct {
	Visibility db.ContactVisibility `json:"visibility" example:"public"`
}

// UpdateContactVisibilityHandler godoc
// @Summary      Update contact visibility
// @Description  update contact visibility
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Param        id   path     int     true  "contact id"
// @Param		 account	   body	   handler.UpdateContactVisibilityRequest	true	"visibility"
// @Success      200  {object}   nil
// @Security     JWT
// @Router       /api/contacts/{id}/visibility [put]
func (tr *transport) UpdateContactVisibilityHandler(c echo.Context) error {
	var data UpdateContactVisibilityRequest
	if err := c.Bind(&data); err != nil {
		return err
	}

	userID := getUserID(c)
	cID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	err := tr.api.UpdateContactVisibility(userID, cID, data.Visibility)

	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// ListMyContactsHandler godoc
// @Summary      List my contacts
// @Description  get my contacts
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Success      200  {object} db.ContactsPage
// @Security     JWT
// @Router       /api/me/contacts [get]
func (tr *transport) ListMyContactsHandler(c echo.Context) error {
	userID := getUserID(c)

	contacts, err := tr.api.ListMyContacts(userID)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, contacts)
}
