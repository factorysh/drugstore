package relationalstore

import (
	"fmt"

	"github.com/factorysh/drugstore/schema"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Store struct {
	db      *sqlx.DB
	schemas map[string]*schema.Schema
}

func New(conn string) (*Store, error) {
	db, err := sqlx.Connect("postgres", conn)
	if err != nil {
		return nil, err
	}
	return &Store{
		db:      db,
		schemas: make(map[string]*schema.Schema),
	}, nil
}

func (s *Store) Register(schema_ *schema.Schema) error {
	sql, err := schema_.DDL()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(sql)
	if err != nil {
		return err
	}
	s.schemas[schema_.Name] = schema_
	return nil
}

func (s *Store) Set(name string, values map[string]interface{}) error {
	schema_, ok := s.schemas[name]
	if !ok {
		return fmt.Errorf("Unknown schema : %s", name)
	}
	sql, err := schema_.Set(values)
	if err != nil {
		return err
	}
	r, err := s.db.NamedExec(sql, values)
	if err != nil {
		return err
	}
	fmt.Println(r)
	return nil
}
