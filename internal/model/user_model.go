package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	Email         string     `db:"email" json:"email"`
	IsActive      bool       `db:"is_active" json:"is_active"`
	EmailVerified bool       `db:"email_verified" json:"email_verified"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

type UserProfile struct {
	UserID      uuid.UUID  `db:"user_id" json:"user_id"`
	Username    string     `db:"username" json:"username"`
	DisplayName *string    `db:"display_name" json:"display_name,omitempty"`
	AvatarURL   *string    `db:"avatar_url" json:"avatar_url,omitempty"`
	Bio         *string    `db:"bio" json:"bio,omitempty"`
	LastLoginAt *time.Time `db:"last_login_at" json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

type UserCredential struct {
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	Password  string    `db:"password" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type UserProvider struct {
	ID             uuid.UUID `db:"id" json:"id"`
	UserID         uuid.UUID `db:"user_id" json:"user_id"`
	Provider       string    `db:"provider" json:"provider"`
	ProviderUserID string    `db:"provider_user_id" json:"provider_user_id"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}
