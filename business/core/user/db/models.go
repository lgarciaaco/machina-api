package db

import (
	"time"

	"github.com/lib/pq"
)

// User represent the structure we need for moving data
// between the app and the database.
type User struct {
	ID             string         `db:"user_id"`
	Name           string         `db:"name"`
	Description    string         `db:"description"`
	Roles          pq.StringArray `db:"roles"`
	PasswordHash   []byte         `db:"password_hash"`
	PositionsTotal int            `db:"positions_total"`
	DateCreated    time.Time      `db:"date_created"`
	DateUpdated    time.Time      `db:"date_updated"`
}
