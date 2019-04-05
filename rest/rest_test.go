package rest

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/factorysh/drugstore/store"

	"github.com/onrik/logrus/filename"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	PROJECT = "project"
)

func newStore() (*store.Store, error) {
	s, err := store.New("postgresql://drugstore:toto@localhost/drugstore?sslmode=disable")
	if err != nil {
		return nil, err
	}
	s.Class(PROJECT, []string{"project", "ns", "name"})
	return s, s.Reset()
}

func rest() (*REST, error) {
	log.SetReportCaller(true)
	filenameHook := filename.NewHook()
	log.AddHook(filenameHook)
	s, err := newStore()
	if err != nil {
		return nil, err
	}
	return &REST{store: s}, nil
}

func TestPost(t *testing.T) {
	r, err := rest()
	assert.NoError(t, err)
	ts := httptest.NewServer(http.HandlerFunc(r.Create))
	defer ts.Close()
	req, err := http.NewRequest("POST", ts.URL+"/project", bytes.NewReader([]byte(`
		{
			"name": "Zoe",
			"ns": "user",
			"project": "drugstore"
		}
	`)))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 201, res.StatusCode)

	rez, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	fmt.Println(string(rez))
}

func TestGet(t *testing.T) {
	r, err := rest()
	assert.NoError(t, err)
	ts := httptest.NewServer(http.HandlerFunc(r.GetByPath))
	defer ts.Close()
	id, err := uuid.NewRandom()
	assert.NoError(t, err)
	err = r.store.Set("project", &store.Document{
		UID: id,
		Data: map[string]interface{}{
			"name":    "yan",
			"ns":      "user",
			"project": "drugstore",
			"age":     42,
			"likes":   []string{"banana", "apple"},
		},
	})
	assert.NoError(t, err)
	resp, err := http.Get(ts.URL + "/project/drugstore/user/yan")
	assert.NoError(t, err)
	rez, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	fmt.Println(string(rez))
}
