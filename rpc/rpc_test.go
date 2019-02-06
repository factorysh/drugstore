package rpc

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/factorysh/drugstore/store"

	"github.com/stretchr/testify/assert"
)

func TestService(t *testing.T) {
	s, err := store.New("postgresql://drugstore:toto@localhost/drugstore?sslmode=disable",
		[]string{"project", "ns", "name"})
	assert.NoError(t, err)
	assert.NotNil(t, s)
	service := New(s)
	id, err := service.Create(map[string]interface{}{
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
