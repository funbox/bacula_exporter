package rdbms

// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Enumerated list of available database providers. Reserved for future.
const (
	POSTGRES = "postgres"
)

// Datastore interface to models
//
// We can then use this interface instead of the direct DB type throughout our application.
// Also we can easily create mock database responses for any unit tests.
//
type Datastore interface {
	GetLatestJobs() ([]*BaculaJob, error)
	GetJobsSummary() ([]*BaculaJobSummary, error)
}

type DB struct {
	*sqlx.DB
}

// ////////////////////////////////////////////////////////////////////////////////// //

// NewDB create new DB struct by datasource connection string with given provider
func NewDB(datasource string) (*DB, error) {
	db, err := sqlx.Open(POSTGRES, datasource)

	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}
