package user

import (
	"time"
)

type UserID int64

type User struct {
	ID UserID

	Username string
	Email    string
	Password string

	RecoveryTokenHash       *string
	RecoveryTokenExpiration *time.Time
	RecoveryTokenUsed       bool
}
