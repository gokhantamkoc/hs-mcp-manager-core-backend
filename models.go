package main

import (
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type OAuthToken struct {
	ID           int       `json:"id,omitempty"`
	UserID       int       `json:"user_id"`
	Provider     string    `json:"provider"`
	AccessToken  string    `json:"-"` 
	RefreshToken string    `json:"-"`
	Expiry       time.Time `json:"expiry"`
}

type Repository struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	Name          string    `json:"name"`
	RepoURL       string    `json:"repo_url"`
	LocalPath     string    `json:"local_path"`
	CurrentStatus string    `json:"current_status"`
	CreatedAt     time.Time `json:"created_at"`
}