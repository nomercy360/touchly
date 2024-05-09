package api

import (
	"touchly/internal/db"
	"touchly/internal/terrors"
)

func (api *api) GetUserByID(userID int64) (*db.User, error) {
	user, err := api.storage.GetUserByID(userID)

	if err != nil {
		if db.IsNoRowsError(err) {
			return nil, terrors.NotFound(err, "user not found")
		} else {
			return nil, terrors.InternalServerError(err, "failed to get user")
		}
	}

	return user, nil
}
