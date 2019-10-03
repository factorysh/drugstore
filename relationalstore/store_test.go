package relationalstore

import (
	"fmt"
	"os"
	"testing"

	"github.com/factorysh/drugstore/schema"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

const (
	SCHEMA = `
---
group:
  type: string
  key: true
project:
  type: string
  key: true
is_drupal:
  type: boolean
plugin_node:
  type: versions
plugin_composer:
  type: versions
`
)

func db() (*Store, error) {
	h := os.Getenv("DB_HOST")
	if h == "" {
		h = "localhost"
	}
	connStr := fmt.Sprintf("postgres://drugstore:toto@%s/drugstore?sslmode=disable", h)
	return New(connStr)
}

func TestStore(t *testing.T) {
	s, err := db()
	assert.NoError(t, err)
	schema_ := schema.Schema{
		Name:   "project",
		Values: make(map[string]schema.Column),
	}
	err = yaml.Unmarshal([]byte(SCHEMA), &schema_.Values)
	assert.NoError(t, err)
	err = s.Register(&schema_)
	assert.NoError(t, err)
}
