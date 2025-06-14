package models

import "time"

type Cookies map[string]string

type Website struct {
	Url         string   `json:"url"`
	CfProtected bool     `json:"cf_protected"`
	Cookies     *Cookies `json:"cookies"`
}

type Session struct {
	IP        string    `json:"ip"`
	UA        string    `json:"ua"`
	Website   *Website  `json:"website"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// just int for now, maybe uuid?
type ID int64

type SessionCache map[ID]Session
