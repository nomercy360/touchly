package handler

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"touchly/internal/db"
)

// ListTagsHandler godoc
// @Summary      List tags
// @Description  list tags
// @Tags         tags
// @Accept       json
// @Produce      json
// @Success      200  {object}   []db.Tag
// @Router       /api/tags [get]
func (tr *transport) ListTagsHandler(c echo.Context) error {
	tags, err := tr.api.ListTags()

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, tags)
}

// CreateTagHandler godoc
// @Summary      Create tag
// @Description  create tag
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        tag body db.Tag true "tag"
// @Success      201  {object}   db.Tag
// @Security     JWT
// @Router       /api/tags [post]
func (tr *transport) CreateTagHandler(c echo.Context) error {
	var tag db.Tag
	if err := c.Bind(&tag); err != nil {
		return err
	}

	createdTag, err := tr.api.CreateTag(tag)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, createdTag)
}

// DeleteTagHandler godoc
// @Summary      Delete tag
// @Description  delete tag
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        id path int true "tag id"
// @Success      200  {object}   nil
// @Security     JWT
// @Router       /api/tags/{id} [delete]
func (tr *transport) DeleteTagHandler(c echo.Context) error {
	id, err := getID(c)
	if err != nil {
		return err
	}

	err = tr.api.DeleteTag(id)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
