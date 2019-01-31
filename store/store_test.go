package store

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {
	s, err := New("postgresql://drugstore:toto@localhost/drugstore?sslmode=disable", []string{"name"})
	assert.NoError(t, err)
	assert.NotNil(t, s)
	fmt.Println(s)
	uid := uuid.MustParse("37AD4002-79A6-4752-A912-AEB111871EBE")
	s.Set(&Document{
		UID: uid,
		Data: map[string]interface{}{
			"name": "Bob",
			"age":  42,
		},
	})
	s.Set(&Document{
		UID: uid,
		Data: map[string]interface{}{
			"name": "Alice",
			"age":  42,
		},
	})

	docs, err := s.GetByUUID(uid)
	assert.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, "Alice", docs[0].Data["name"])

}
