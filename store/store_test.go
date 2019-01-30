package store

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {
	s, err := New("postgresql://drugstore:toto@localhost/drugstore?sslmode=disable")
	assert.NoError(t, err)
	assert.NotNil(t, s)
	fmt.Println(s)
	s.Set(&Document{
		UID: uuid.MustParse("37AD4002-79A6-4752-A912-AEB111871EBE"),
		Data: json.RawMessage([]byte(`{
			"name": "Bob",
			"age": 42
		}
			`)),
	})
	s.Set(&Document{
		UID: uuid.MustParse("37AD4002-79A6-4752-A912-AEB111871EBE"),
		Data: json.RawMessage([]byte(`{
			"name": "Alice",
			"age": 42
		}
			`)),
	})
}
