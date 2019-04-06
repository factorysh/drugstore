package store

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"

	jmespath "github.com/jmespath/go-jmespath"

	"encoding/json"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Store store things
type Store struct {
	db        *sqlx.DB
	paths     map[string][]string // class => paths
	parser    *jmespath.Parser
	startPath *regexp.Regexp
}

var schema = `
CREATE TABLE IF NOT EXISTS document (
	uid         UUID UNIQUE,
	mtime		TIMESTAMP NOT NULL,
	ctime       TIMESTAMP NOT NULL,
	class       TEXT NOT NULL,
	data        JSONB,
	PRIMARY KEY (uid)
  );
CREATE INDEX IF NOT EXISTS idxginp ON document USING GIN (data jsonb_path_ops);
`

// New Store
func New(dsn string) (*Store, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}
	return &Store{
		db:        db,
		paths:     make(map[string][]string),
		parser:    jmespath.NewParser(),
		startPath: regexp.MustCompile("^[a-zA-Z0-9:_.]+"),
	}, nil
}

func (s *Store) Class(name string, paths []string) {
	s.paths[name] = paths
}

// Set Document (create or update)
func (s *Store) Set(class string, d *Document) error {
	paths, ok := s.paths[class]
	if !ok {
		return fmt.Errorf("Unknown class : %s", class)
	}
	for _, path := range paths {
		_, ok := d.Data[path]
		if !ok {
			return fmt.Errorf("Key %s is mandatory", path)
		}
	}
	dd, err := json.Marshal(d.Data)
	if err != nil {
		return err
	}
	if d.UID == nil {
		u, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		d.UID = &u
	}
	tx := s.db.MustBegin()
	tx.MustExec(`
	INSERT INTO  document AS d (uid, class, data, ctime, mtime)
	VALUES ($1, $2, $3, $4, $4)
	ON CONFLICT (uid) DO UPDATE
	SET data=$3, mtime=$4 WHERE d.uid=$1`, d.UID.String(), class, dd, time.Now())
	tx.Commit()
	return nil
}

func (s *Store) byLabel(docs []Document, key string) (map[string][]Document, error) {
	data := make(map[string][]Document)
	for _, doc := range docs {
		v, ok := doc.Data[key]
		if !ok {
			return nil, fmt.Errorf("Key not found : %s", key)
		}
		vv, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("Value is not a string : %s => %s", key, v)
		}
		_, ok = data[vv]
		if !ok {
			data[vv] = []Document{doc}
		} else {
			data[vv] = append(data[vv], doc)
		}
	}
	return data, nil
}

func recurse(data map[string]interface{}, keys []string) (map[string]interface{}, []string) {
	key := keys[0]
	var m map[string]interface{}
	var ok bool
	mm, ok := data[key]
	if !ok {
		m = make(map[string]interface{})
		data[key] = m
	} else {
		m, ok = mm.(map[string]interface{})
		if !ok {
			panic("Unknow type")
		}
	}
	if len(keys) > 1 {
		return recurse(m, keys[1:])
	}
	return m, []string{}
}

// Documents2tree build a tree from a collection of documents
func (s *Store) Documents2tree(class string, docs []Document) (map[string]interface{}, error) {
	tree := make(map[string]interface{})
	for _, doc := range docs {
		err := s.tree(class, tree, doc.Data)
		if err != nil {
			return nil, err
		}
	}
	return tree, nil
}

func (s *Store) tree(class string, data map[string]interface{}, doc map[string]interface{}) error {
	paths := s.paths[class]
	keys := make([]string, len(paths))
	for i, path := range paths {
		v, ok := doc[path]
		if !ok {

		}
		vv, ok := v.(string)

		keys[i] = vv
	}
	leaf, _ := recurse(data, keys[:len(keys)-1])
	leaf[keys[len(keys)-1]] = doc

	return nil
}

func (s *Store) Delete(uid uuid.UUID) error {
	_, err := s.db.Queryx("DELETE FROM document WHERE uid=$1", uid.String())
	return err
}

func (s *Store) Length() (int, error) {
	rows, err := s.db.Queryx("SELECT COUNT(*) AS count FROM document")
	if err != nil {
		return 0, err
	}
	rows.Next()
	var l int
	err = rows.Rows.Scan(&l)
	return l, err
}

func (s *Store) Reset() error {
	_, err := s.db.Queryx("DELETE FROM document")
	return err
}
