package domain

import (
	"errors"
	"time"
	"tp25-api/lib"
)

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleMonitor Role = "monitor" // readonly
)

type User struct {
	ID       string   `json:"id" bson:"_id"`
	Username string   `json:"username" bson:"username"`
	FullName string   `json:"full_name" bson:"full_name"`
	Role     Role     `json:"role" bson:"role"`
	Phone    string   `json:"phone" bson:"phone"`
	ZoneID   *string  `json:"zone_id,omitempty" bson:"zone_id,omitempty"`
	Groups   []string `json:"groups" bson:"groups"`
	ZaloID   *string  `json:"zalo_id,omitempty" bson:"zalo_id,omitempty"`
	CTime    int64    `json:"ctime" bson:"ctime"`
	MTime    int64    `json:"mtime" bson:"mtime"`
	DTime    *int64   `json:"dtime,omitempty" bson:"dtime,omitempty"`
}

type CreateUserParams struct {
	Username string   `json:"username" binding:"required"`
	FullName string   `json:"full_name" binding:"required"`
	Role     Role     `json:"role" binding:"required"`
	Phone    string   `json:"phone"`
	Groups   []string `json:"groups"`
	ZaloID   *string  `json:"zalo_id"`
}

type UpdateUserParams struct {
	FullName *string  `json:"full_name"`
	Phone    *string  `json:"phone"`
	Groups   []string `json:"groups"`
	ZaloID   *string  `json:"zalo_id"`
}

type UserSecret struct {
	UserID string `json:"user_id" bson:"user_id"`
	Name   string `json:"name" bson:"name"`
	Value  string `json:"value" bson:"value"`
	Encode string `json:"encode" bson:"encode"`
}

type RefreshToken struct {
	ID        string `json:"id" bson:"_id"`
	UserID    string `json:"user_id" bson:"user_id"`
	ExpiresAt int64  `json:"expires_at" bson:"expires_at"`
	CTime     int64  `json:"ctime" bson:"ctime"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SetPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUsernameNotFound    = errors.New("username not found")
	ErrUsernameExisted     = errors.New("username existed")
	ErrUserHasNoLogin      = errors.New("user has no login")
	ErrWrongPassword       = errors.New("wrong password")
	ErrInvalidSession      = errors.New("invalid session")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrUnauthorized        = errors.New("unauthorized")
)

// NewUser creates a new user with timestamps
func NewUser(params CreateUserParams) *User {
	now := time.Now().UnixMilli()
	return &User{
		ID:       lib.Rand.Char(12),
		Username: params.Username,
		FullName: params.FullName,
		Role:     params.Role,
		Phone:    params.Phone,
		Groups:   params.Groups,
		ZaloID:   params.ZaloID,
		CTime:    now,
		MTime:    now,
	}
}
