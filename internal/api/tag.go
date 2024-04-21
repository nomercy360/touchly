package api

import (
	"touchly/internal/db"
	"touchly/internal/terrors"
)

func (api *api) ListTags() ([]db.Tag, error) {
	tags, err := api.storage.ListTags()

	if err != nil {
		return nil, terrors.InternalServerError(err, "failed to list tags")
	}

	return tags, nil
}

func (api *api) CreateTag(tag db.Tag) (*db.Tag, error) {
	if tag.Name == "" {
		return nil, terrors.InvalidRequest(nil, "name is required")
	}

	res, err := api.storage.CreateTag(tag)

	if err != nil {
		return nil, terrors.InternalServerError(err, "failed to create tag")
	}

	return res, nil
}

func (api *api) DeleteTag(id int64) error {
	if err := api.storage.DeleteTag(id); err != nil {
		return terrors.InternalServerError(err, "failed to delete tag")
	}

	return nil
}
