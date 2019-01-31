package store

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"encoding/json"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type RawDocument struct {
	UID  uuid.UUID       `db:"uid"`
	Data json.RawMessage `db:"data"`
}

type Document struct {
	UID  uuid.UUID
	Data map[string]interface{}
}

// Store store things
type Store struct {
	db    *sqlx.DB
	paths []string
}

var schema = `
CREATE TABLE IF NOT EXISTS document (
	uid         UUID UNIQUE,
	data        JSONB,
	PRIMARY KEY (uid)
  );
CREATE INDEX IF NOT EXISTS idxginp ON document USING GIN (data jsonb_path_ops);
`

// New Store
func New(dsn string, paths []string) (*Store, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}
	return &Store{
		db:    db,
		paths: paths,
	}, nil
}

// Set Document
func (s *Store) Set(d *Document) error {
	for _, path := range s.paths {
		_, ok := d.Data[path]
		if !ok {
			return fmt.Errorf("Key %s is mandatory", path)
		}
	}
	dd, err := json.Marshal(d.Data)
	if err != nil {
		return err
	}
	tx := s.db.MustBegin()
	tx.MustExec(`
	INSERT INTO  document AS d (uid, data) VALUES ($1, $2)
	ON CONFLICT (uid ) DO UPDATE
	SET data=$2 WHERE d.uid=$1`, d.UID.String(), dd)
	tx.Commit()
	return nil
}

func (s *Store) GetByUUID(uuids ...uuid.UUID) ([]Document, error) {
	var documents []RawDocument
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
	// FIXME it's ugly, sqlx job is done twice
	ds := make([]Document, len(documents))
	for i, d := range documents {
		var dd map[string]interface{}
		err = json.Unmarshal(d.Data, &dd)
		if err != nil {
			return nil, err
		}
		ds[i] = Document{
			UID:  d.UID,
			Data: dd,
		}
	}
	return ds, nil
}
