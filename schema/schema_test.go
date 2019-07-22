package schema

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"encoding/json"

	_ "github.com/lib/pq"

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

func TestDDL(t *testing.T) {
	schema, err := New("project", []byte(SCHEMA))
	assert.NoError(t, err)
	ddl, err := schema.DDL()
	assert.NoError(t, err)
	fmt.Println(ddl)
	if !testing.Short() {
		db, err := db()
		assert.NoError(t, err)
		rows, err := db.Query(ddl)
		assert.NoError(t, err)
		fmt.Println(rows)
	}
}

func TestSet(t *testing.T) {
	schema, err := New("project", []byte(SCHEMA))
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
	sql, err := schema.Set(data)
	assert.NoError(t, err)
	fmt.Println(sql)
	if !testing.Short() {
		db, err := db()
		assert.NoError(t, err)
		rows, err := db.Query(sql)
		fmt.Println(rows)
	}
	assert.False(t, true)
}
