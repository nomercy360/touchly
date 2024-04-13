package db

import (
	"database/sql"
	"errors"
	"github.com/lib/pq"
)

func IsNoRowsError(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func IsDuplicationError(err error) bool {
	if err == nil {
		return false
	}

	var pqErr *pq.Error
	ok := errors.As(err, &pqErr)
	return ok && pqErr.Code == "23505"
}

func IsForeignKeyViolationError(err error) bool {
	if err == nil {
		return false
	}

	var pqErr *pq.Error
	ok := errors.As(err, &pqErr)
	return ok && pqErr.Code == "23503"
}
