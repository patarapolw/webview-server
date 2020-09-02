package sqlite

import (
	"github.com/jmoiron/sqlx"
)

// StatementStore storage of SQL statements
type StatementStore struct {
	DB    *sqlx.DB
	store map[string](*sqlx.Stmt)
}

// Get a statement from store. Return nil if not exists
func (s *StatementStore) Get(id string) *sqlx.Stmt {
	return s.store[id]
}

// Free free up all sql statements
func (s *StatementStore) Free() {
	for id, stmt := range s.store {
		stmt.Close()
		s.store[id] = nil
	}
}
