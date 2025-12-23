package model

import (
	"database/sql"
)

type User struct {
	ID                      int64
	Username                string
	Password                string
	Email                   string

	RecoveryTokenHash       sql.NullString
	RecoveryTokenExpiration sql.NullTime
	RecoveryTokenUsed       bool
}
