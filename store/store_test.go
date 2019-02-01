package store

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {
	s, err := New("postgresql://drugstore:toto@localhost/drugstore?sslmode=disable",
		[]string{"project", "ns", "name"})
	assert.NoError(t, err)
	assert.NotNil(t, s)
	fmt.Println(s)
	uid := uuid.MustParse("37AD4002-79A6-4752-A912-AEB111871EBE")
	err = s.Set(&Document{
		UID: uid,
		Data: map[string]interface{}{
			"project": "drugstore",
			"name":    "Bob",
			"ns":      "user",
			"age":     42,
		},
	})
	assert.NoError(t, err)
	err = s.Set(&Document{
		UID: uid,
		Data: map[string]interface{}{
			"project": "drugstore",
			"name":    "Alice",
			"ns":      "user",
			"age":     42,
		},
	})
	assert.NoError(t, err)

	err = s.Set(&Document{
		UID: uuid.MustParse("BBED4C33-3925-4E56-A806-A75A7BAB46A9"),
		Data: map[string]interface{}{
			"project": "drugstore",
			"ns":      "user",
			"name":    "Charle",
			"age":     18,
		},
	})
	assert.NoError(t, err)

	docs, err := s.GetByUUID(uid)
	assert.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, "Alice", docs[0].Data["name"])

	docs, err = s.GetByPath("drugstore")
	assert.NoError(t, err)
	assert.Len(t, docs, 2)

	docs, err = s.GetByJMEspath("drugstore.Alice.*")
	assert.NoError(t, err)

	all, err := s.GetByPath()
	assert.NoError(t, err)
	names, err := s.byLabel(all, "name")
	assert.NoError(t, err)
	spew.Dump(names)

	tree := make(map[string]interface{})
	for _, a := range all {
		err = s.tree(tree, a.Data)
		assert.NoError(t, err)
	}
	spew.Dump(tree)
}
