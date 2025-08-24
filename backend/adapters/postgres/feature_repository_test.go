package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/feature-voting-platform/backend/domain/features"
	"github.com/feature-voting-platform/backend/domain/votes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeatureRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewFeatureRepository(&DB{db})
	now := time.Now()

	tests := []struct {
		name    string
		feature *features.Feature
		setup   func()
		wantErr bool
	}{
		{
			name: "successful creation",
			feature: &features.Feature{
				Title:       "Test Feature",
				Description: "Test Description",
				CreatedBy:   1,
			},
			setup: func() {
				mock.ExpectQuery(`INSERT INTO features`).
					WithArgs("Test Feature", "Test Description", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "vote_count", "created_at", "updated_at"}).
						AddRow(1, 0, now, now))
			},
			wantErr: false,
		},
		{
			name: "database error",
			feature: &features.Feature{
				Title:       "Test Feature",
				Description: "Test Description",
				CreatedBy:   1,
			},
			setup: func() {
				mock.ExpectQuery(`INSERT INTO features`).
					WithArgs("Test Feature", "Test Description", 1).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := repo.Create(tt.feature)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 1, tt.feature.ID)
				assert.Equal(t, 0, tt.feature.VoteCount)
				assert.Equal(t, now, tt.feature.CreatedAt)
				assert.Equal(t, now, tt.feature.UpdatedAt)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFeatureRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewFeatureRepository(&DB{db})
	now := time.Now()

	tests := []struct {
		name    string
		id      int
		userID  *int
		setup   func()
		want    *features.Feature
		wantErr bool
	}{
		{
			name:   "feature found without user vote check",
			id:     1,
			userID: nil,
			setup: func() {
				mock.ExpectQuery(`SELECT f.id, f.title, f.description, f.created_by, u.username, f.vote_count, f.created_at, f.updated_at FROM features f LEFT JOIN users u ON f.created_by = u.id WHERE f.id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "created_by", "username", "vote_count", "created_at", "updated_at"}).
						AddRow(1, "Test Feature", "Test Description", 1, "testuser", 5, now, now))
			},
			want: &features.Feature{
				ID:              1,
				Title:           "Test Feature",
				Description:     "Test Description",
				CreatedBy:       1,
				CreatedByUser:   stringPtr("testuser"),
				VoteCount:       5,
				CreatedAt:       now,
				UpdatedAt:       now,
				HasUserVoted:    false,
			},
			wantErr: false,
		},
		{
			name:   "feature found with user vote check - has voted",
			id:     1,
			userID: intPtr(2),
			setup: func() {
				mock.ExpectQuery(`SELECT f.id, f.title, f.description, f.created_by, u.username, f.vote_count, f.created_at, f.updated_at FROM features f LEFT JOIN users u ON f.created_by = u.id WHERE f.id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "created_by", "username", "vote_count", "created_at", "updated_at"}).
						AddRow(1, "Test Feature", "Test Description", 1, "testuser", 5, now, now))

				mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM votes WHERE user_id = \$1 AND feature_id = \$2\)`).
					WithArgs(2, 1).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
			want: &features.Feature{
				ID:              1,
				Title:           "Test Feature",
				Description:     "Test Description",
				CreatedBy:       1,
				CreatedByUser:   stringPtr("testuser"),
				VoteCount:       5,
				CreatedAt:       now,
				UpdatedAt:       now,
				HasUserVoted:    true,
			},
			wantErr: false,
		},
		{
			name:   "feature not found",
			id:     999,
			userID: nil,
			setup: func() {
				mock.ExpectQuery(`SELECT f.id, f.title, f.description, f.created_by, u.username, f.vote_count, f.created_at, f.updated_at FROM features f LEFT JOIN users u ON f.created_by = u.id WHERE f.id = \$1`).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			feature, err := repo.GetByID(tt.id, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, feature)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, feature)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFeatureRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewFeatureRepository(&DB{db})
	now := time.Now()

	tests := []struct {
		name     string
		page     int
		perPage  int
		userID   *int
		setup    func()
		want     []features.Feature
		wantTotal int
		wantErr  bool
	}{
		{
			name:    "successful retrieval without user",
			page:    1,
			perPage: 10,
			userID:  nil,
			setup: func() {
				// Mock count query
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM features`).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

				// Mock features query
				mock.ExpectQuery(`SELECT f.id, f.title, f.description, f.created_by, u.username, f.vote_count, f.created_at, f.updated_at FROM features f LEFT JOIN users u ON f.created_by = u.id ORDER BY f.created_at DESC LIMIT \$1 OFFSET \$2`).
					WithArgs(10, 0).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "created_by", "username", "vote_count", "created_at", "updated_at"}).
						AddRow(1, "Feature 1", "Description 1", 1, "user1", 3, now, now).
						AddRow(2, "Feature 2", "Description 2", 2, "user2", 1, now, now))
			},
			want: []features.Feature{
				{
					ID:              1,
					Title:           "Feature 1",
					Description:     "Description 1",
					CreatedBy:       1,
					CreatedByUser:   stringPtr("user1"),
					VoteCount:       3,
					CreatedAt:       now,
					UpdatedAt:       now,
					HasUserVoted:    false,
				},
				{
					ID:              2,
					Title:           "Feature 2",
					Description:     "Description 2",
					CreatedBy:       2,
					CreatedByUser:   stringPtr("user2"),
					VoteCount:       1,
					CreatedAt:       now,
					UpdatedAt:       now,
					HasUserVoted:    false,
				},
			},
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:    "count query error",
			page:    1,
			perPage: 10,
			userID:  nil,
			setup: func() {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM features`).
					WillReturnError(sql.ErrConnDone)
			},
			want:      nil,
			wantTotal: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			features, total, err := repo.GetAll(tt.page, tt.perPage, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, features)
				assert.Equal(t, 0, total)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, features)
				assert.Equal(t, tt.wantTotal, total)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFeatureRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewFeatureRepository(&DB{db})

	tests := []struct {
		name        string
		id          int
		title       *string
		description *string
		setup       func()
		wantErr     bool
	}{
		{
			name:        "update title only",
			id:          1,
			title:       stringPtr("Updated Title"),
			description: nil,
			setup: func() {
				mock.ExpectExec(`UPDATE features SET title = \$1 WHERE id = \$2`).
					WithArgs("Updated Title", 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:        "update both title and description",
			id:          1,
			title:       stringPtr("Updated Title"),
			description: stringPtr("Updated Description"),
			setup: func() {
				mock.ExpectExec(`UPDATE features SET title = \$1, description = \$2 WHERE id = \$3`).
					WithArgs("Updated Title", "Updated Description", 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:        "no fields to update",
			id:          1,
			title:       nil,
			description: nil,
			setup:       func() {},
			wantErr:     true,
		},
		{
			name:        "feature not found",
			id:          999,
			title:       stringPtr("Updated Title"),
			description: nil,
			setup: func() {
				mock.ExpectExec(`UPDATE features SET title = \$1 WHERE id = \$2`).
					WithArgs("Updated Title", 999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := repo.Update(tt.id, tt.title, tt.description)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFeatureRepository_FeatureExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewFeatureRepository(&DB{db})

	tests := []struct {
		name    string
		id      int
		setup   func()
		want    bool
		wantErr bool
	}{
		{
			name: "feature exists",
			id:   1,
			setup: func() {
				mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM features WHERE id = \$1\)`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "feature does not exist",
			id:   999,
			setup: func() {
				mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM features WHERE id = \$1\)`).
					WithArgs(999).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "database error",
			id:   1,
			setup: func() {
				mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM features WHERE id = \$1\)`).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			exists, err := repo.FeatureExists(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.False(t, exists)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, exists)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFeatureRepository_AddVote(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewFeatureRepository(&DB{db})

	tests := []struct {
		name      string
		userID    int
		featureID int
		setup     func()
		wantErr   bool
	}{
		{
			name:      "successful vote addition",
			userID:    1,
			featureID: 1,
			setup: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO votes \(user_id, feature_id\) VALUES \(\$1, \$2\)`).
					WithArgs(1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:      "database error",
			userID:    1,
			featureID: 1,
			setup: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO votes \(user_id, feature_id\) VALUES \(\$1, \$2\)`).
					WithArgs(1, 1).
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := repo.AddVote(tt.userID, tt.featureID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFeatureRepository_HasUserVoted(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewFeatureRepository(&DB{db})

	tests := []struct {
		name      string
		userID    int
		featureID int
		setup     func()
		want      bool
		wantErr   bool
	}{
		{
			name:      "user has voted",
			userID:    1,
			featureID: 1,
			setup: func() {
				mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM votes WHERE user_id = \$1 AND feature_id = \$2\)`).
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
			want:    true,
			wantErr: false,
		},
		{
			name:      "user has not voted",
			userID:    1,
			featureID: 1,
			setup: func() {
				mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM votes WHERE user_id = \$1 AND feature_id = \$2\)`).
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			hasVoted, err := repo.HasUserVoted(tt.userID, tt.featureID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.False(t, hasVoted)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, hasVoted)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFeatureRepository_GetUserVotes(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewFeatureRepository(&DB{db})
	now := time.Now()

	tests := []struct {
		name    string
		userID  int
		setup   func()
		want    []votes.Vote
		wantErr bool
	}{
		{
			name:   "successful retrieval",
			userID: 1,
			setup: func() {
				mock.ExpectQuery(`SELECT v.id, v.user_id, v.feature_id, v.created_at FROM votes v WHERE v.user_id = \$1 ORDER BY v.created_at DESC`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "feature_id", "created_at"}).
						AddRow(1, 1, 10, now).
						AddRow(2, 1, 20, now))
			},
			want: []votes.Vote{
				{ID: 1, UserID: 1, FeatureID: 10, CreatedAt: now},
				{ID: 2, UserID: 1, FeatureID: 20, CreatedAt: now},
			},
			wantErr: false,
		},
		{
			name:   "no votes found",
			userID: 1,
			setup: func() {
				mock.ExpectQuery(`SELECT v.id, v.user_id, v.feature_id, v.created_at FROM votes v WHERE v.user_id = \$1 ORDER BY v.created_at DESC`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "feature_id", "created_at"}))
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			votes, err := repo.GetUserVotes(tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, votes)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, votes)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}