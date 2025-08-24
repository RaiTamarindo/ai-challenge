package models

import (
	"time"
)

type Feature struct {
	ID              int       `json:"id" db:"id"`
	Title           string    `json:"title" db:"title"`
	Description     string    `json:"description" db:"description"`
	CreatedBy       int       `json:"created_by" db:"created_by"`
	CreatedByUser   string    `json:"created_by_username,omitempty" db:"created_by_username"`
	VoteCount       int       `json:"vote_count" db:"vote_count"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	HasUserVoted    bool      `json:"has_user_voted,omitempty"`
}

type CreateFeatureRequest struct {
	Title       string `json:"title" binding:"required,min=5,max=255"`
	Description string `json:"description" binding:"required,min=10"`
}

type UpdateFeatureRequest struct {
	Title       *string `json:"title,omitempty" binding:"omitempty,min=5,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,min=10"`
}

type Vote struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	FeatureID int       `json:"feature_id" db:"feature_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type VoteRequest struct {
	FeatureID int `json:"feature_id" binding:"required"`
}

type FeatureListResponse struct {
	Features []Feature `json:"features"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PerPage  int       `json:"per_page"`
}