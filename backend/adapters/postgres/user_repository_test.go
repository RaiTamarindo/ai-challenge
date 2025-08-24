package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/feature-voting-platform/backend/domain/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(&DB{db})
	now := time.Now()

	tests := []struct {
		name    string
		user    *users.User
		setup   func()
		wantErr bool
	}{
		{
			name: "successful creation",
			user: &users.User{
				Username:     "testuser",
				Email:        "test@example.com",
				PasswordHash: "hashed_password",
			},
			setup: func() {
				mock.ExpectQuery(`INSERT INTO users`).
					WithArgs("testuser", "test@example.com", "hashed_password").
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
						AddRow(1, now, now))
			},
			wantErr: false,
		},
		{
			name: "database error",
			user: &users.User{
				Username:     "testuser",
				Email:        "test@example.com",
				PasswordHash: "hashed_password",
			},
			setup: func() {
				mock.ExpectQuery(`INSERT INTO users`).
					WithArgs("testuser", "test@example.com", "hashed_password").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := repo.Create(tt.user)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 1, tt.user.ID)
				assert.Equal(t, now, tt.user.CreatedAt)
				assert.Equal(t, now, tt.user.UpdatedAt)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(&DB{db})
	now := time.Now()

	tests := []struct {
		name    string
		email   string
		setup   func()
		want    *users.User
		wantErr bool
	}{
		{
			name:  "user found",
			email: "test@example.com",
			setup: func() {
				mock.ExpectQuery(`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE email = \$1`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "created_at", "updated_at"}).
						AddRow(1, "testuser", "test@example.com", "hashed_password", now, now))
			},
			want: &users.User{
				ID:           1,
				Username:     "testuser",
				Email:        "test@example.com",
				PasswordHash: "hashed_password",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			wantErr: false,
		},
		{
			name:  "user not found",
			email: "nonexistent@example.com",
			setup: func() {
				mock.ExpectQuery(`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE email = \$1`).
					WithArgs("nonexistent@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:  "database error",
			email: "test@example.com",
			setup: func() {
				mock.ExpectQuery(`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE email = \$1`).
					WithArgs("test@example.com").
					WillReturnError(sql.ErrConnDone)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			user, err := repo.GetByEmail(tt.email)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, user)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(&DB{db})
	now := time.Now()

	tests := []struct {
		name    string
		id      int
		setup   func()
		want    *users.User
		wantErr bool
	}{
		{
			name: "user found",
			id:   1,
			setup: func() {
				mock.ExpectQuery(`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "created_at", "updated_at"}).
						AddRow(1, "testuser", "test@example.com", "hashed_password", now, now))
			},
			want: &users.User{
				ID:           1,
				Username:     "testuser",
				Email:        "test@example.com",
				PasswordHash: "hashed_password",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			wantErr: false,
		},
		{
			name: "user not found",
			id:   999,
			setup: func() {
				mock.ExpectQuery(`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE id = \$1`).
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

			user, err := repo.GetByID(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, user)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(&DB{db})
	now := time.Now()

	tests := []struct {
		name     string
		username string
		setup    func()
		want     *users.User
		wantErr  bool
	}{
		{
			name:     "user found",
			username: "testuser",
			setup: func() {
				mock.ExpectQuery(`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE username = \$1`).
					WithArgs("testuser").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "created_at", "updated_at"}).
						AddRow(1, "testuser", "test@example.com", "hashed_password", now, now))
			},
			want: &users.User{
				ID:           1,
				Username:     "testuser",
				Email:        "test@example.com",
				PasswordHash: "hashed_password",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			wantErr: false,
		},
		{
			name:     "user not found",
			username: "nonexistent",
			setup: func() {
				mock.ExpectQuery(`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE username = \$1`).
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			user, err := repo.GetByUsername(tt.username)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, user)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(&DB{db})
	now := time.Now()

	tests := []struct {
		name    string
		user    *users.User
		setup   func()
		wantErr bool
	}{
		{
			name: "successful update",
			user: &users.User{
				ID:           1,
				Username:     "updated_user",
				Email:        "updated@example.com",
				PasswordHash: "new_hashed_password",
			},
			setup: func() {
				mock.ExpectQuery(`UPDATE users SET username = \$2, email = \$3, password_hash = \$4, updated_at = CURRENT_TIMESTAMP WHERE id = \$1`).
					WithArgs(1, "updated_user", "updated@example.com", "new_hashed_password").
					WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow(now))
			},
			wantErr: false,
		},
		{
			name: "database error",
			user: &users.User{
				ID:           1,
				Username:     "updated_user",
				Email:        "updated@example.com",
				PasswordHash: "new_hashed_password",
			},
			setup: func() {
				mock.ExpectQuery(`UPDATE users SET username = \$2, email = \$3, password_hash = \$4, updated_at = CURRENT_TIMESTAMP WHERE id = \$1`).
					WithArgs(1, "updated_user", "updated@example.com", "new_hashed_password").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := repo.Update(tt.user)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, now, tt.user.UpdatedAt)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(&DB{db})

	tests := []struct {
		name    string
		id      int
		setup   func()
		wantErr bool
	}{
		{
			name: "successful deletion",
			id:   1,
			setup: func() {
				mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "user not found",
			id:   999,
			setup: func() {
				mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "database error",
			id:   1,
			setup: func() {
				mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := repo.Delete(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}