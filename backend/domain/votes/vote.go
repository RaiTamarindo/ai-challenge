package votes

import (
	"time"
)

// Vote represents the core vote entity
type Vote struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	FeatureID int       `json:"feature_id"`
	CreatedAt time.Time `json:"created_at"`
}

// VoteRequest represents the data needed to cast a vote
type VoteRequest struct {
	FeatureID int `json:"feature_id" binding:"required"`
}