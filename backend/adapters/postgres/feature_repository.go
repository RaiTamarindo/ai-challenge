package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/feature-voting-platform/backend/domain/features"
	"github.com/feature-voting-platform/backend/domain/votes"
)

// FeatureRepository implements both features.Repository and votes.Repository interfaces
type FeatureRepository struct {
	db *DB
}

// NewFeatureRepository creates a new feature repository
func NewFeatureRepository(db *DB) *FeatureRepository {
	return &FeatureRepository{db: db}
}

// Feature-related methods implementing features.Repository

// Create creates a new feature in the database
func (r *FeatureRepository) Create(feature *features.Feature) error {
	query := `
		INSERT INTO features (title, description, created_by)
		VALUES ($1, $2, $3)
		RETURNING id, vote_count, created_at, updated_at
	`
	
	err := r.db.QueryRow(query, feature.Title, feature.Description, feature.CreatedBy).
		Scan(&feature.ID, &feature.VoteCount, &feature.CreatedAt, &feature.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create feature: %w", err)
	}
	
	return nil
}

// GetByID retrieves a feature by ID
func (r *FeatureRepository) GetByID(id int, userID *int) (*features.Feature, error) {
	feature := &features.Feature{}
	query := `
		SELECT f.id, f.title, f.description, f.created_by, u.username,
		       f.vote_count, f.created_at, f.updated_at
		FROM features f
		LEFT JOIN users u ON f.created_by = u.id
		WHERE f.id = $1
	`
	
	err := r.db.QueryRow(query, id).Scan(
		&feature.ID, &feature.Title, &feature.Description, &feature.CreatedBy,
		&feature.CreatedByUser, &feature.VoteCount, &feature.CreatedAt, &feature.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("feature not found")
		}
		return nil, fmt.Errorf("failed to get feature by ID: %w", err)
	}
	
	// Check if user has voted for this feature
	if userID != nil {
		hasVoted, err := r.HasUserVoted(*userID, id)
		if err != nil {
			return nil, fmt.Errorf("failed to check user vote status: %w", err)
		}
		feature.HasUserVoted = hasVoted
	}
	
	return feature, nil
}

// GetAll retrieves all features with pagination
func (r *FeatureRepository) GetAll(page, perPage int, userID *int) ([]features.Feature, int, error) {
	offset := (page - 1) * perPage
	
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM features`
	err := r.db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get features count: %w", err)
	}
	
	// Get features with pagination, sorted by vote count (most voted first)
	query := `
		SELECT f.id, f.title, f.description, f.created_by, u.username,
		       f.vote_count, f.created_at, f.updated_at
		FROM features f
		LEFT JOIN users u ON f.created_by = u.id
		ORDER BY f.vote_count DESC, f.created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := r.db.Query(query, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get features: %w", err)
	}
	defer rows.Close()
	
	var featuresList []features.Feature
	for rows.Next() {
		var feature features.Feature
		err := rows.Scan(
			&feature.ID, &feature.Title, &feature.Description, &feature.CreatedBy,
			&feature.CreatedByUser, &feature.VoteCount, &feature.CreatedAt, &feature.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan feature: %w", err)
		}
		
		// Check if user has voted for this feature
		if userID != nil {
			hasVoted, err := r.HasUserVoted(*userID, feature.ID)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to check user vote status: %w", err)
			}
			feature.HasUserVoted = hasVoted
		}
		
		featuresList = append(featuresList, feature)
	}
	
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating features: %w", err)
	}
	
	return featuresList, total, nil
}

// GetByCreatedBy retrieves features created by a specific user
func (r *FeatureRepository) GetByCreatedBy(userID int) ([]features.Feature, error) {
	query := `
		SELECT f.id, f.title, f.description, f.created_by, u.username,
		       f.vote_count, f.created_at, f.updated_at
		FROM features f
		LEFT JOIN users u ON f.created_by = u.id
		WHERE f.created_by = $1
		ORDER BY f.created_at DESC
	`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get features by user: %w", err)
	}
	defer rows.Close()
	
	var featuresList []features.Feature
	for rows.Next() {
		var feature features.Feature
		err := rows.Scan(
			&feature.ID, &feature.Title, &feature.Description, &feature.CreatedBy,
			&feature.CreatedByUser, &feature.VoteCount, &feature.CreatedAt, &feature.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feature: %w", err)
		}
		featuresList = append(featuresList, feature)
	}
	
	return featuresList, nil
}

// Update updates a feature
func (r *FeatureRepository) Update(id int, title, description *string) error {
	setParts := []string{}
	args := []interface{}{}
	argCount := 1
	
	if title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argCount))
		args = append(args, *title)
		argCount++
	}
	
	if description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argCount))
		args = append(args, *description)
		argCount++
	}
	
	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}
	
	query := fmt.Sprintf("UPDATE features SET %s WHERE id = $%d", 
		strings.Join(setParts, ", "), argCount)
	args = append(args, id)
	
	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update feature: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("feature not found")
	}
	
	return nil
}

// Delete deletes a feature
func (r *FeatureRepository) Delete(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Delete votes first
	_, err = tx.Exec(`DELETE FROM votes WHERE feature_id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete votes: %w", err)
	}
	
	// Delete feature
	result, err := tx.Exec(`DELETE FROM features WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete feature: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("feature not found")
	}
	
	return tx.Commit()
}

// FeatureExists checks if a feature exists
func (r *FeatureRepository) FeatureExists(id int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM features WHERE id = $1)`
	
	err := r.db.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if feature exists: %w", err)
	}
	
	return exists, nil
}

// Vote-related methods implementing votes.Repository

// AddVote adds a vote for a feature
func (r *FeatureRepository) AddVote(userID, featureID int) error {
	// Begin transaction with SERIALIZABLE isolation level
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Set transaction isolation level to SERIALIZABLE
	_, err = tx.Exec("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE")
	if err != nil {
		return fmt.Errorf("failed to set isolation level: %w", err)
	}
	
	// Insert vote
	query := `INSERT INTO votes (user_id, feature_id) VALUES ($1, $2)`
	_, err = tx.Exec(query, userID, featureID)
	if err != nil {
		return fmt.Errorf("failed to add vote: %w", err)
	}
	
	// Update feature vote count
	updateQuery := `UPDATE features SET vote_count = vote_count + 1 WHERE id = $1`
	_, err = tx.Exec(updateQuery, featureID)
	if err != nil {
		return fmt.Errorf("failed to update vote count: %w", err)
	}
	
	return tx.Commit()
}

// RemoveVote removes a vote from a feature
func (r *FeatureRepository) RemoveVote(userID, featureID int) error {
	// Begin transaction with SERIALIZABLE isolation level
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Set transaction isolation level to SERIALIZABLE
	_, err = tx.Exec("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE")
	if err != nil {
		return fmt.Errorf("failed to set isolation level: %w", err)
	}
	
	// Delete vote
	query := `DELETE FROM votes WHERE user_id = $1 AND feature_id = $2`
	result, err := tx.Exec(query, userID, featureID)
	if err != nil {
		return fmt.Errorf("failed to remove vote: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("vote not found")
	}
	
	// Update feature vote count (decrement)
	updateQuery := `UPDATE features SET vote_count = vote_count - 1 WHERE id = $1`
	_, err = tx.Exec(updateQuery, featureID)
	if err != nil {
		return fmt.Errorf("failed to update vote count: %w", err)
	}
	
	return tx.Commit()
}

// HasUserVoted checks if a user has voted for a feature
func (r *FeatureRepository) HasUserVoted(userID, featureID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM votes WHERE user_id = $1 AND feature_id = $2)`
	
	err := r.db.QueryRow(query, userID, featureID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user vote: %w", err)
	}
	
	return exists, nil
}

// GetUserVotes retrieves all votes made by a user
func (r *FeatureRepository) GetUserVotes(userID int) ([]votes.Vote, error) {
	query := `
		SELECT v.id, v.user_id, v.feature_id, v.created_at
		FROM votes v
		WHERE v.user_id = $1
		ORDER BY v.created_at DESC
	`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user votes: %w", err)
	}
	defer rows.Close()
	
	var votesList []votes.Vote
	for rows.Next() {
		var vote votes.Vote
		err := rows.Scan(
			&vote.ID, &vote.UserID, &vote.FeatureID, &vote.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vote: %w", err)
		}
		votesList = append(votesList, vote)
	}
	
	return votesList, nil
}