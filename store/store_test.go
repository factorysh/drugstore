package store

import (
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
)

const (
	PROJECT = "project"
)

var (
	ALICE  = uuid.MustParse("37AD4002-79A6-4752-A912-AEB111871EBE")
	BOB    = uuid.MustParse("BBED4C33-3925-4E56-A806-A75A7BAB46A9")
	CHARLY = uuid.MustParse("E12C6B1F-EBEA-4987-AFD8-C880B43FD963")
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
	err = s.Set(PROJECT, &Document{
		UID: &BOB,
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
		UID: &ALICE,
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
	err = s.Set(PROJECT, &Document{
		UID: &CHARLY,
		Data: map[string]interface{}{
			"project": "drugstore",
			"name":    "Charly",
			"ns":      "user",
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
	assert.Equal(t, 3, l)
}

func TestByUUID(t *testing.T) {
	s, err := fixture()
	assert.NoError(t, err)

	docs, err := s.GetByUUID(ALICE)
	assert.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, "Alice", docs[0].Data["name"])
}

func TestByPath(t *testing.T) {
	s, err := fixture()
	assert.NoError(t, err)

	docs, err := s.GetByPath(PROJECT, "drugstore")
	assert.NoError(t, err)
	assert.Len(t, docs, 3)

	docs, err = s.GetByPath(PROJECT, "", "user")
	assert.NoError(t, err)
	assert.Len(t, docs, 3)

	docs, err = s.GetByPath(PROJECT, "", "", "Charly")
	assert.NoError(t, err)
	assert.Len(t, docs, 1)

	all, err := s.GetByPath(PROJECT)
	assert.NoError(t, err)
	names, err := s.byLabel(all, "name")
	assert.NoError(t, err)
	spew.Dump(names)

	tree, err := s.Documents2tree(PROJECT, all)
	assert.NoError(t, err)
	spew.Dump(tree)
}

func TestByJMespath(t *testing.T) {
	s, err := fixture()
	assert.NoError(t, err)

	resp, err := s.GetByJMEspath(PROJECT, "*.user.*[]|[?age>`18`]")
	assert.NoError(t, err)
	spew.Dump(resp)
	r, ok := resp.([]interface{})
	assert.True(t, ok)
	assert.Len(t, r, 2)
	//assert.Equal(t, "Alice", r[0]["name"])
}

func TestDelete(t *testing.T) {
	s, err := fixture()
	assert.NoError(t, err)

	err = s.Delete(uuid.MustParse("BBED4C33-3925-4E56-A806-A75A7BAB46A9"))
	assert.NoError(t, err)
	l, err := s.Length()
	assert.NoError(t, err)
	assert.Equal(t, 2, l)
}

func TestSet(t *testing.T) {
	s, err := fixture()
	assert.NoError(t, err)
	docs := []RawDocument{}
	err = s.db.Select(&docs, "SELECT * FROM document")
	assert.NoError(t, err)
	for _, doc := range docs {
		fmt.Println(string(doc.Data))
	}
	fmt.Println("docs: ", docs)
	l, err := s.Length()
	assert.NoError(t, err)
	assert.Equal(t, 3, l)
	err = s.Set(PROJECT, &Document{
		Data: map[string]interface{}{
			"project": "drugstore",
			"ns":      "user",
			"name":    "Charly",
			"age":     21,
		},
	})
	assert.NoError(t, err)
	l, err = s.Length()
	assert.NoError(t, err)
	assert.Equal(t, 3, l)
}
