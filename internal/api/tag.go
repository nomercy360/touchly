package api

import (
	"errors"
	"touchly/internal/db"
)

func (api *api) ListTags() ([]db.Tag, error) {
	tags, err := api.storage.ListTags()

	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (api *api) CreateTag(tag db.Tag) (*db.Tag, error) {
	if tag.Name == "" {
		return nil, errors.New("name is required")
	}

	res, err := api.storage.CreateTag(tag)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (api *api) DeleteTag(id int64) error {
	if err := api.storage.DeleteTag(id); err != nil {
		return err
	}

	return nil
}
