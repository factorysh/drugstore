package schema

import (
	"fmt"
	"testing"

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
	assert.True(t, false)
}
