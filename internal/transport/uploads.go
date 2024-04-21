package transport

import (
	"net/http"
)

// GetUploadURLHandler godoc
// @Summary      Get upload URL
// @Description  returns a presigned URL for uploading a file
// @Tags         uploads
// @Accept       json
// @Produce      json
// @Param        file_name query string true "file name"
// @Success      200  {object}   api.UploadURL
// @Security     JWT
// @Router       /api/uploads/get-url [post]
func (tr *transport) GetUploadURLHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file_name")
	userID := getUserIDFromRequest(r)

	res, err := tr.api.GetPresignedURL(userID, fileName)

	if err != nil {
		WriteError(r, w, err)
		return
	}

	WriteJSON(w, http.StatusOK, res)
}
