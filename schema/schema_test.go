package schema

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"encoding/json"

	_ "github.com/lib/pq"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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

func db() (*sql.DB, error) {
	h := os.Getenv("DB_HOST")
	if h == "" {
		h = "localhost"
	}
	connStr := fmt.Sprintf("postgres://drugstore:toto@%s/drugstore?sslmode=disable", h)
	return sql.Open("postgres", connStr)
}

func TestDB(t *testing.T) {
	log.SetLevel(log.InfoLevel)
	schema, err := New("project", []byte(SCHEMA))
	assert.NoError(t, err)
	db, err := db()
	assert.NoError(t, err)
	d := NewDB(db, schema)
	err = d.Create()
	assert.NoError(t, err)
	var data map[string]interface{}
	err = json.Unmarshal([]byte(`{
"group": "factory",
"project": "drugstore",
"is_drupal": false,
"plugin_node": {
	"requests" : "2.88.0"
}
}`), &data)
	assert.NoError(t, err)
	err = d.Upsert(data)
	assert.NoError(t, err)
	err = json.Unmarshal([]byte(`{
"group": "factory",
"project": "drugstore",
"is_drupal": false,
"plugin_node": {
	"jquery" : "7",
	"requests" : "2.88.1"
}
}`), &data)
	assert.NoError(t, err)
	err = d.Upsert(data)
	assert.NoError(t, err)
	err = d.Delete(map[string]interface{}{
		"group":   "factory",
		"project": "drugstore",
	})
	assert.NoError(t, err)
}
