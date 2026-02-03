package model

import (
	"time"
)

type User struct {
	ID       int64
	Username string
	Password string
	Email    string

	RecoveryTokenHash       string
	RecoveryTokenExpiration *time.Time
	RecoveryTokenUsed       bool
}
