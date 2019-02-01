package store

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	jmespath "github.com/jmespath/go-jmespath"

	"encoding/json"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
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

func (s *Store) GetByPath(paths ...string) ([]Document, error) {
	if len(paths) > len(s.paths) {
		return nil, fmt.Errorf("Path too long : %s", paths)
	}
	buf := bytes.NewBuffer([]byte(`
		SELECT *
		FROM document
	`))
	if len(paths) > 0 {
		buf.WriteString("WHERE ")

		for i, p := range paths {
			buf.WriteString(` data @> '{"`)
			buf.WriteString(s.paths[i])
			buf.WriteString(`": "`)
			buf.WriteString(p)
			buf.WriteString(`"}'`)
			if i+1 < len(paths) {
				buf.WriteString(" AND ")
			}
		}
	}
	l := log.WithField("sql", buf.String())
	var documents []RawDocument
	err := s.db.Select(&documents, buf.String())
	if err != nil {
		l.WithError(err).Error("GetByPath")
		return nil, err
	}
	l.Info("GetByPath")
	docs := make([]Document, len(documents))
	for i, d := range documents {
		var dd map[string]interface{}
		err := json.Unmarshal(d.Data, &dd)
		if err != nil {
			l.WithError(err).Error("GetByPath")
			return nil, err
		}
		docs[i] = Document{
			UID:  d.UID,
			Data: dd,
		}
	}
	return docs, nil
}

func (s *Store) GetByJMEspath(path string) ([]Document, error) {
	/*
		n, err := s.parser.Parse(path)
		if err != nil {
			return nil, err
		}
		fmt.Println(n)
	*/
	p := s.startPath.FindString(path)
	var docs []Document
	var err error
	if p != "" {
		if strings.HasSuffix(p, ".") {
			p = p[:len(p)-1]
		}
		pp := strings.Split(p, ".")
		docs, err = s.GetByPath(pp...)
		if err != nil {
			return nil, err
		}
	}
	data := make(map[string]interface{})
	for _, d := range docs {
		for _, p := range s.paths {
			k, ok := d.Data[p]
			if !ok {
				return nil, fmt.Errorf("Can't find key %s", p)
			}
			fmt.Println(k)
			kk, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("Key is not a string : %s => %s", p, k)
			}
			data[kk] = d
		}
	}
	jm, err := jmespath.Search(path, data)
	if err != nil {
		return nil, err
	}
	fmt.Println(jm)
	fmt.Println(data)
	return nil, nil
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
