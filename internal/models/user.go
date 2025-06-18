package models

import "time"

// this is more than likely going to change
type (
	Password  [16]byte
	SessionID [16]byte
)

type Token struct {
	ID        ID        `json:"id"`
	String    string    `json:"string"`
	ExpiresAt time.Time `json:"expires_at"`
}

type User struct {
	ID          ID        `json:"id"`
	Name        string    `json:"name"`
	Password    Password  `json:"password"`
	Admin       bool      `json:"admin"`
	Token       Token     `json:"token"`
	CreatedAt   time.Time `json:"created_at"`
	LastUpdated time.Time `json:"last_updated"`
}

type UserSession struct {
	SessionID   SessionID `json:"session_id"`
	UserID      ID        `json:"user_id"`
	IP          string    `json:"ip"`
	UA          string    `json:"ua"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	LastUpdated time.Time `json:"last_login"`
}
