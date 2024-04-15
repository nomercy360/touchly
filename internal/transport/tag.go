package transport

import (
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
func (tr *transport) ListTagsHandler(w http.ResponseWriter, r *http.Request) {
	tags, err := tr.api.ListTags()

	if err != nil {
		_ = WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = WriteJSON(w, http.StatusOK, tags)
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
func (tr *transport) CreateTagHandler(w http.ResponseWriter, r *http.Request) {
	var tag db.Tag
	if err := decodeRequest(r, &tag); err != nil {
		_ = WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	createdTag, err := tr.api.CreateTag(tag)
	if err != nil {
		_ = WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = WriteJSON(w, http.StatusCreated, createdTag)
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
func (tr *transport) DeleteTagHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		_ = WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = tr.api.DeleteTag(id)
	if err != nil {
		_ = WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = WriteJSON(w, http.StatusOK, nil)
}
