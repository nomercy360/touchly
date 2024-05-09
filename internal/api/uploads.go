package api

import (
	"fmt"
	"time"
	"touchly/internal/terrors"
)

type UploadURL struct {
	URL string `json:"url"`
} // @Name UploadURL

func (api *api) GetPresignedURL(userID int64, fileName string) (*UploadURL, error) {
	var res UploadURL

	if fileName == "" {
		return &res, terrors.InvalidRequest(nil, "file_name is required")
	}

	fileName = fmt.Sprintf("%d/%s", userID, fileName)

	url, err := api.s3Client.GetPresignedURL(fileName, 15*time.Minute)

	if err != nil {
		return &res, terrors.InternalServerError(err, "failed to get presigned URL")
	}

	return &UploadURL{URL: url}, nil
}
