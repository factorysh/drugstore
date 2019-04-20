package conf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConf(t *testing.T) {
	cfg, err := New([]byte(`
---
postgresql_dsn: postgresql://drugstore:toto@localhost/drugstore?sslmode=disable
classes:
  project:
   - project
   - ns
   - name 
`))
	assert.NoError(t, err)
	assert.Len(t, cfg.Classes["project"], 3)
}
