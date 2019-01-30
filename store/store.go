package store

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Document struct {
	ID   int             `db:"id"`
	UID  uuid.UUID       `db:"uid"`
	Data json.RawMessage `db:"data"`
}

// Store store things
type Store struct {
	db *sqlx.DB
}

var schema = `
CREATE TABLE IF NOT EXISTS document (
	uid         UUID UNIQUE,
	data        JSONB,
	PRIMARY KEY (uid)
  );
`

// New Store
func New(dsn string) (*Store, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.MustExec(schema)
	return &Store{
		db: db,
	}, nil
}

// Set Document
func (s *Store) Set(d *Document) {
	fmt.Println(s.db)
	tx := s.db.MustBegin()
	tx.MustExec(`
	INSERT INTO  document AS d (uid, data) VALUES ($1, $2)
	ON CONFLICT (uid ) DO UPDATE
	SET data=$2 WHERE d.uid=$1`, d.UID.String(), d.Data)
	tx.Commit()
}

func (s *Store) GetByUUID(uuids ...uuid.UUID) ([]Document, error) {
	documents := []Document{}
	ids := make([]string, len(uuids))
	for i, uid := range uuids {
		// FIXME: OMG it's ugly!
		ids[i] = "'" + uid.String() + "'"
	}
	err := s.db.Select(&documents, `
		SELECT *
		FROM document
		WHERE uid IN (`+strings.Join(ids, ",")+")")
	if err != nil {
		return nil, err
	}
	return documents, nil
}
