package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	ts := httptest.NewServer(http.HandlerFunc(r.Handler))
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
	ts := httptest.NewServer(http.HandlerFunc(r.Handler))
	defer ts.Close()
	id, err := uuid.NewRandom()
	assert.NoError(t, err)
	err = r.store.Set("project", &store.Document{
		UID: &id,
		Data: map[string]interface{}{
			"name":    "yann",
			"ns":      "user",
			"project": "drugstore",
			"age":     42,
			"likes":   []string{"banana", "apple"},
		},
	})
	assert.NoError(t, err)

	id, err = uuid.NewRandom()
	assert.NoError(t, err)
	err = r.store.Set("project", &store.Document{
		UID: &id,
		Data: map[string]interface{}{
			"name":    "walter",
			"ns":      "user",
			"project": "drugstore",
			"age":     23,
			"likes":   []string{"orange"},
		},
	})
	assert.NoError(t, err)
	type responses []map[string]interface{}

	resp, err := http.Get(ts.URL + "/project/drugstore/user/yann")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	rez, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	fmt.Println(string(rez))
	var rs responses
	err = json.Unmarshal(rez, &rs)
	assert.NoError(t, err)
	assert.Len(t, rs, 1)
	assert.Equal(t, "yann", rs[0]["name"])

	resp, err = http.Get(ts.URL + "/project/drugstore/user/xavier")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	rez, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	fmt.Println(string(rez))
	err = json.Unmarshal(rez, &rs)
	assert.NoError(t, err)
	assert.Len(t, rs, 0)

	resp, err = http.Get(ts.URL + "/project/drugstore/user/")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	rez, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	fmt.Println(string(rez))
	err = json.Unmarshal(rez, &rs)
	assert.NoError(t, err)
	assert.Len(t, rs, 2)

}

func TestHome(t *testing.T) {
	r := New(nil)
	ts := httptest.NewServer(http.HandlerFunc(r.Handler))
	defer ts.Close()
	resp, err := http.DefaultClient.Get(ts.URL)
	assert.NoError(t, err)
	rez, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	fmt.Println(string(rez))
}

func TestQuery(t *testing.T) {
	r, err := rest()
	assert.NoError(t, err)
	ts := httptest.NewServer(http.HandlerFunc(r.Handler))
	defer ts.Close()

	err = r.store.Set("project", &store.Document{
		Data: map[string]interface{}{
			"name":    "yann",
			"ns":      "user",
			"project": "drugstore",
			"age":     42,
			"likes":   []string{"banana", "apple"},
		},
	})
	assert.NoError(t, err)

	err = r.store.Set("project", &store.Document{
		Data: map[string]interface{}{
			"name":    "walter",
			"ns":      "user",
			"project": "drugstore",
			"age":     23,
			"likes":   []string{"orange"},
		},
	})
	assert.NoError(t, err)
	type responses []map[string]interface{}

	var rs responses
	resp, err := http.DefaultClient.Get(ts.URL + "/project/_search?q=" +
		url.QueryEscape("*.user.*[]|[?name=='walter']"))
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	rez, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	fmt.Println(string(rez))
	err = json.Unmarshal(rez, &rs)
	assert.NoError(t, err)
	assert.Len(t, rs, 1)
	assert.Equal(t, float64(23), rs[0]["age"])
}
