package rpc

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/factorysh/drugstore/store"

	"github.com/stretchr/testify/assert"
)

const (
	PROJECT = "project"
)

func TestService(t *testing.T) {
	s, err := store.New("postgresql://drugstore:toto@localhost/drugstore?sslmode=disable")
	assert.NoError(t, err)
	assert.NotNil(t, s)
	s.Class(PROJECT, []string{"project", "ns", "name"})
	service := New(s)
	id, err := service.Create(PROJECT, map[string]interface{}{
		"project": "drugstore",
		"ns":      "user",
		"name":    "bob",
	})
	assert.NoError(t, err)
	uid, err := uuid.Parse(id)
	assert.NoError(t, err)
	fmt.Println(uid)
	d, err := service.Get(id)
	fmt.Println(d)
}
