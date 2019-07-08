package schema

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"

	"github.com/stretchr/testify/assert"
)

func TestSchema(t *testing.T) {

	schema, err := New([]byte(`
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
`))
	assert.NoError(t, err)
	ddl, err := schema.DDL("project")
	assert.NoError(t, err)
	fmt.Println(ddl)
	if !testing.Short() {
		h := os.Getenv("DB_HOST")
		if h == "" {
			h = "localhost"
		}
		connStr := fmt.Sprintf("postgres://drugstore:toto@%s/drugstore?sslmode=disable", h)
		db, err := sql.Open("postgres", connStr)
		assert.NoError(t, err)
		fmt.Println(ddl)
		rows, err := db.Query(ddl)
		assert.NoError(t, err)
		fmt.Println(rows)
	}
}
