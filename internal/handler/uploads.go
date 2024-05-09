package handler

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// GetUploadURLHandler godoc
// @Summary      Get upload URL
// @Description  returns a presigned URL for uploading a file
// @Tags         uploads
// @Accept       json
// @Produce      json
// @Param        file_name query string true "file name"
// @Success      200  {object}  UploadURL
// @Security     JWT
// @Router       /api/uploads/get-url [post]
func (tr *transport) GetUploadURLHandler(c echo.Context) error {
	fileName := c.QueryParam("file_name")
	userID := getUserID(c)

	res, err := tr.api.GetPresignedURL(userID, fileName)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}
