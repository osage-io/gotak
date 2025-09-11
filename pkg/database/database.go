package database

import (
	"context"
	"database/sql"
)

var (
	// ErrNoRows is returned when a query returns no rows
	ErrNoRows = sql.ErrNoRows
)

// DB represents a database connection interface
type DB interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (Result, error)
	BeginTx(ctx context.Context, opts *TxOptions) (Tx, error)
}

// Tx represents a database transaction interface
type Tx interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (Result, error)
	Commit() error
	Rollback() error
}

// Row represents a database row interface
type Row interface {
	Scan(dest ...interface{}) error
}

// Rows represents multiple database rows interface
type Rows struct {
	*sql.Rows
}

// Result represents a database result interface
type Result interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

// TxOptions represents transaction options
type TxOptions struct {
	Isolation sql.IsolationLevel
	ReadOnly  bool
}

// Wrapper implementations for sql.DB
type SQLDatabase struct {
	*sql.DB
}

// NewSQLDatabase creates a new SQL database wrapper
func NewSQLDatabase(db *sql.DB) *SQLDatabase {
	return &SQLDatabase{DB: db}
}

func (db *SQLDatabase) QueryContext(ctx context.Context, query string, args ...interface{}) (*Rows, error) {
	rows, err := db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{rows}, nil
}

func (db *SQLDatabase) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	return db.DB.QueryRowContext(ctx, query, args...)
}

func (db *SQLDatabase) ExecContext(ctx context.Context, query string, args ...interface{}) (Result, error) {
	return db.DB.ExecContext(ctx, query, args...)
}

func (db *SQLDatabase) BeginTx(ctx context.Context, opts *TxOptions) (Tx, error) {
	var txOpts *sql.TxOptions
	if opts != nil {
		txOpts = &sql.TxOptions{
			Isolation: opts.Isolation,
			ReadOnly:  opts.ReadOnly,
		}
	}
	
	tx, err := db.DB.BeginTx(ctx, txOpts)
	if err != nil {
		return nil, err
	}
	
	return &SQLTransaction{tx}, nil
}

// SQLTransaction wraps sql.Tx
type SQLTransaction struct {
	*sql.Tx
}

func (tx *SQLTransaction) QueryContext(ctx context.Context, query string, args ...interface{}) (*Rows, error) {
	rows, err := tx.Tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{rows}, nil
}

func (tx *SQLTransaction) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	return tx.Tx.QueryRowContext(ctx, query, args...)
}

func (tx *SQLTransaction) ExecContext(ctx context.Context, query string, args ...interface{}) (Result, error) {
	return tx.Tx.ExecContext(ctx, query, args...)
}

// IsConnectionError checks if the error is a database connection error
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for common connection errors
	errStr := err.Error()
	return contains(errStr, "connection refused") ||
		contains(errStr, "connection reset") ||
		contains(errStr, "connection lost") ||
		contains(errStr, "network is unreachable")
}

// IsConstraintError checks if the error is a constraint violation
func IsConstraintError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	return contains(errStr, "constraint") ||
		contains(errStr, "duplicate key") ||
		contains(errStr, "unique violation")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr || 
			 indexSubstring(s, substr) >= 0)))
}

func indexSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}
	
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
