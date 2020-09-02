package sqlite

import (
	"github.com/jmoiron/sqlx"
)

// BodyQuery request body of how it should queried
type BodyQuery struct {
	ID     string        `json:"id"` // For stored statement
	SQL    string        `json:"sql" binding:"required"`
	Params []interface{} `json:"params"`
}

// Prepare returns prepared statement
// If there is ID, add request body to the store
func (body *BodyQuery) Prepare(s *StatementStore) *sqlx.Stmt {
	if body.SQL == "" {
		if body.ID == "" {
			panic("body.ID not supplied")
		}

		return s.store[body.ID]
	}

	stmt, err := s.DB.Preparex(body.SQL)
	if err != nil {
		panic(err)
	}

	if body.ID != "" {
		if s.store[body.ID] != nil {
			s.store[body.ID].Close()
		}

		s.store[body.ID] = stmt
	}

	return stmt
}
