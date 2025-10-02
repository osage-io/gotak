package database

import (
	"context"
	"database/sql"
	
	"github.com/jmoiron/sqlx"
)

// SQLXAdapter wraps sqlx.DB to implement the DB interface
type SQLXAdapter struct {
	db *sqlx.DB
}

// NewSQLXAdapter creates a new SQLX adapter
func NewSQLXAdapter(db *sqlx.DB) *SQLXAdapter {
	return &SQLXAdapter{db: db}
}

// QueryContext implements DB interface
func (s *SQLXAdapter) QueryContext(ctx context.Context, query string, args ...interface{}) (*Rows, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{rows}, nil
}

// QueryRowContext implements DB interface
func (s *SQLXAdapter) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	return s.db.QueryRowContext(ctx, query, args...)
}

// ExecContext implements DB interface
func (s *SQLXAdapter) ExecContext(ctx context.Context, query string, args ...interface{}) (Result, error) {
	return s.db.ExecContext(ctx, query, args...)
}

// BeginTx implements DB interface
func (s *SQLXAdapter) BeginTx(ctx context.Context, opts *TxOptions) (Tx, error) {
	var txOpts *sql.TxOptions
	if opts != nil {
		txOpts = &sql.TxOptions{
			Isolation: opts.Isolation,
			ReadOnly:  opts.ReadOnly,
		}
	}
	
	tx, err := s.db.BeginTxx(ctx, txOpts)
	if err != nil {
		return nil, err
	}
	
	return &SQLXTransaction{tx}, nil
}

// SQLXTransaction wraps sqlx.Tx
type SQLXTransaction struct {
	tx *sqlx.Tx
}

// QueryContext implements Tx interface
func (t *SQLXTransaction) QueryContext(ctx context.Context, query string, args ...interface{}) (*Rows, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{rows}, nil
}

// QueryRowContext implements Tx interface
func (t *SQLXTransaction) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

// ExecContext implements Tx interface
func (t *SQLXTransaction) ExecContext(ctx context.Context, query string, args ...interface{}) (Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

// Commit implements Tx interface
func (t *SQLXTransaction) Commit() error {
	return t.tx.Commit()
}

// Rollback implements Tx interface
func (t *SQLXTransaction) Rollback() error {
	return t.tx.Rollback()
}