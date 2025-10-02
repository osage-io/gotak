package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrUserExists is returned when trying to create a user that already exists
	ErrUserExists = errors.New("user already exists")
	// ErrSessionNotFound is returned when a session is not found
	ErrSessionNotFound = errors.New("session not found")
)

// Repository provides database operations for users
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new user repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new user in the database
func (r *Repository) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (
			id, username, email, password_hash, role, callsign, active,
			created_at, updated_at
		) VALUES (
			:id, :username, :email, :password_hash, :role, :callsign, :active,
			:created_at, :updated_at
		)`

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Active = true

	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return ErrUserExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	query := `
		SELECT id, username, email, password_hash, role, callsign, active,
		       created_at, updated_at, last_login
		FROM users
		WHERE id = $1`

	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *Repository) GetByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	query := `
		SELECT id, username, email, password_hash, role, callsign, active,
		       created_at, updated_at, last_login
		FROM users
		WHERE LOWER(username) = LOWER($1)`

	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	query := `
		SELECT id, username, email, password_hash, role, callsign, active,
		       created_at, updated_at, last_login
		FROM users
		WHERE LOWER(email) = LOWER($1)`

	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// Update updates a user in the database
func (r *Repository) Update(ctx context.Context, user *User) error {
	query := `
		UPDATE users
		SET username = :username,
		    email = :email,
		    role = :role,
		    callsign = :callsign,
		    active = :active,
		    updated_at = :updated_at
		WHERE id = :id`

	user.UpdatedAt = time.Now()

	result, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdatePassword updates a user's password
func (r *Repository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1,
		    updated_at = $2
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, passwordHash, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateLastLogin updates a user's last login time
func (r *Repository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET last_login = $1
		WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// Delete deletes a user from the database
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// List retrieves a list of users with optional filtering
func (r *Repository) List(ctx context.Context, filter *UserFilter) ([]*User, int, error) {
	var users []*User
	var total int

	// Build the WHERE clause
	whereClauses := []string{}
	args := []interface{}{}
	argCount := 0

	if filter.Active != nil {
		argCount++
		whereClauses = append(whereClauses, fmt.Sprintf("active = $%d", argCount))
		args = append(args, *filter.Active)
	}

	if filter.Role != nil {
		argCount++
		whereClauses = append(whereClauses, fmt.Sprintf("role = $%d", argCount))
		args = append(args, *filter.Role)
	}

	if filter.SearchTerm != nil && *filter.SearchTerm != "" {
		argCount++
		searchPattern := "%" + *filter.SearchTerm + "%"
		whereClauses = append(whereClauses, fmt.Sprintf(
			"(username ILIKE $%d OR email ILIKE $%d OR callsign ILIKE $%d)",
			argCount, argCount, argCount,
		))
		args = append(args, searchPattern)
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Build the main query
	orderBy := filter.OrderBy
	if orderBy == "" {
		orderBy = "created_at"
	}

	orderDir := filter.OrderDir
	if orderDir == "" {
		orderDir = "DESC"
	}

	query := fmt.Sprintf(`
		SELECT id, username, email, password_hash, role, callsign, active,
		       created_at, updated_at, last_login
		FROM users
		%s
		ORDER BY %s %s
		LIMIT %d OFFSET %d`,
		whereClause, orderBy, orderDir, filter.Limit, filter.Offset,
	)

	err = r.db.SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// CreateSession creates a new session
func (r *Repository) CreateSession(ctx context.Context, session *Session) error {
	query := `
		INSERT INTO sessions (
			id, user_id, token_hash, expires_at, ip_address, user_agent, created_at
		) VALUES (
			:id, :user_id, :token_hash, :expires_at, :ip_address, :user_agent, :created_at
		)`

	session.ID = uuid.New()
	session.CreatedAt = time.Now()

	_, err := r.db.NamedExecContext(ctx, query, session)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetSessionByTokenHash retrieves a session by token hash
func (r *Repository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*Session, error) {
	var session Session
	query := `
		SELECT id, user_id, token_hash, expires_at, ip_address, user_agent, created_at
		FROM sessions
		WHERE token_hash = $1 AND expires_at > $2`

	err := r.db.GetContext(ctx, &session, query, tokenHash, time.Now())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// DeleteSession deletes a session
func (r *Repository) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// DeleteUserSessions deletes all sessions for a user
func (r *Repository) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions from the database
func (r *Repository) CleanupExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < $1`

	_, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	return nil
}

// SaveLoginAttempt records a login attempt
func (r *Repository) SaveLoginAttempt(ctx context.Context, attempt *LoginAttempt) error {
	query := `
		INSERT INTO login_attempts (
			id, username, ip_address, success, attempted_at
		) VALUES (
			:id, :username, :ip_address, :success, :attempted_at
		)`

	attempt.ID = uuid.New()
	attempt.AttemptedAt = time.Now()

	_, err := r.db.NamedExecContext(ctx, query, attempt)
	if err != nil {
		return fmt.Errorf("failed to save login attempt: %w", err)
	}

	return nil
}

// GetRecentLoginAttempts retrieves recent login attempts for rate limiting
func (r *Repository) GetRecentLoginAttempts(ctx context.Context, ipAddress string, since time.Time) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM login_attempts
		WHERE ip_address = $1 AND attempted_at > $2 AND success = false`

	err := r.db.GetContext(ctx, &count, query, ipAddress, since)
	if err != nil {
		return 0, fmt.Errorf("failed to count login attempts: %w", err)
	}

	return count, nil
}