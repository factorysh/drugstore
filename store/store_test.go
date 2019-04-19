package store

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
)

const (
	PROJECT = "project"
)

func fixture() (*Store, error) {
	s, err := New("postgresql://drugstore:toto@localhost/drugstore?sslmode=disable")
	if err != nil {
		return nil, err
	}
	s.Class(PROJECT, []string{"project", "ns", "name"})
	err = s.Reset()
	if err != nil {
		return nil, err
	}
	uid := uuid.MustParse("37AD4002-79A6-4752-A912-AEB111871EBE")
	err = s.Set(PROJECT, &Document{
		UID: &uid,
		Data: map[string]interface{}{
			"project": "drugstore",
			"name":    "Bob",
			"ns":      "user",
			"age":     42,
		},
	})
	if err != nil {
		return nil, err
	}
	err = s.Set(PROJECT, &Document{
		UID: &uid,
		Data: map[string]interface{}{
			"project": "drugstore",
			"name":    "Alice",
			"ns":      "user",
			"age":     42,
		},
	})
	if err != nil {
		return nil, err
	}
	u := uuid.MustParse("BBED4C33-3925-4E56-A806-A75A7BAB46A9")
	err = s.Set(PROJECT, &Document{
		UID: &u,
		Data: map[string]interface{}{
			"project": "drugstore",
			"ns":      "user",
			"name":    "Charle",
			"age":     18,
		},
	})
	if err != nil {
		return nil, err
	}
	return s, nil
}

func TestStore(t *testing.T) {
	s, err := fixture()
	assert.NoError(t, err)

	l, err := s.Length()
	assert.NoError(t, err)
	assert.Equal(t, 2, l)

	uid := uuid.MustParse("37AD4002-79A6-4752-A912-AEB111871EBE")
	docs, err := s.GetByUUID(uid)
	assert.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, "Alice", docs[0].Data["name"])

	docs, err = s.GetByPath(PROJECT, "drugstore")
	assert.NoError(t, err)
	assert.Len(t, docs, 2)

	docs, err = s.GetByPath(PROJECT, "", "user")
	assert.NoError(t, err)
	assert.Len(t, docs, 2)

	docs, err = s.GetByPath(PROJECT, "", "", "Charle")
	assert.NoError(t, err)
	assert.Len(t, docs, 1)

	resp, err := s.GetByJMEspath(PROJECT, "*.user.*[]|[?age>`18`]")
	assert.NoError(t, err)
	spew.Dump(resp)
	r, ok := resp.([]interface{})
	assert.True(t, ok)
	assert.Len(t, r, 1)
	//assert.Equal(t, "Alice", r[0]["name"])

	all, err := s.GetByPath(PROJECT)
	assert.NoError(t, err)
	names, err := s.byLabel(all, "name")
	assert.NoError(t, err)
	spew.Dump(names)

	tree, err := s.Documents2tree(PROJECT, all)
	assert.NoError(t, err)
	spew.Dump(tree)

	err = s.Delete(uuid.MustParse("BBED4C33-3925-4E56-A806-A75A7BAB46A9"))
	assert.NoError(t, err)
	l, err = s.Length()
	assert.NoError(t, err)
	assert.Equal(t, 1, l)
}
