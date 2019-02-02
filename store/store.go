package store

import (
	"fmt"
	"regexp"

	jmespath "github.com/jmespath/go-jmespath"

	"encoding/json"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Store store things
type Store struct {
	db        *sqlx.DB
	paths     []string
	parser    *jmespath.Parser
	startPath *regexp.Regexp
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
		db:        db,
		paths:     paths,
		parser:    jmespath.NewParser(),
		startPath: regexp.MustCompile("^[a-zA-Z0-9:_.]+"),
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
	dd, err = json.Marshal(d.Data)
	if err != nil {
		return err
	}
	tx.MustExec(`
	INSERT INTO  document AS d (uid, data) VALUES ($1, $2)
	ON CONFLICT (uid) DO UPDATE
	SET data=$2 WHERE d.uid=$1`, d.UID.String(), dd)
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
func (s *Store) Documents2tree(docs []Document) (map[string]interface{}, error) {
	tree := make(map[string]interface{})
	for _, doc := range docs {
		err := s.tree(tree, doc.Data)
		if err != nil {
			return nil, err
		}
	}
	return tree, nil
}

func (s *Store) tree(data map[string]interface{}, doc map[string]interface{}) error {
	keys := make([]string, len(s.paths))
	for i, path := range s.paths {
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
