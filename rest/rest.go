package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/factorysh/drugstore/store"
	log "github.com/sirupsen/logrus"
)

type REST struct {
	store *store.Store
}

// GetByPath get an object, from its path
func (rest *REST) GetByPath(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	l := log.WithField("url", r.URL.String())
	slugs := strings.Split(r.RequestURI, "/")[1:]
	if len(slugs) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	class := slugs[0]
	slugs = slugs[1:]
	docs, err := rest.store.GetByPath(class, slugs...)
	if err != nil {
		l.WithError(err).Error()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf := bytes.NewBuffer([]byte("[\n"))
	for i, doc := range docs {
		l.WithField("doc", doc)
		j, err := json.Marshal(doc.Data)
		if err != nil {
			l.WithError(err).Error()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		buf.Write(j)
		if i+1 < len(docs) {
			buf.WriteString(",\n")
		}
	}
	buf.WriteString("\n]")
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes())
}

// Query  GET /{slugs}?q={query}
func (rest *REST) Query(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	slugs := strings.Split(r.URL.Path, "/")
	l := log.WithField("url", r.URL.String()).WithField("slugs", slugs)
	if len(slugs) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	class := slugs[1]
	q := r.URL.Query().Get("q")
	l.WithField("class", class).WithField("q", q)
	resp, err := rest.store.GetByJMEspath(class, q)
	if err != nil {
		l.WithError(err).Error()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	j, err := json.Marshal(resp)
	if err != nil {
		l.WithError(err).Error()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func (rest *REST) Create(w http.ResponseWriter, r *http.Request) {
	l := log.WithField("url", r.URL.String()).WithField("method", r.Method)
	if r.Method != "POST" {
		l.Error("Create")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	slugs := strings.Split(r.RequestURI, "/")[1:]
	l = l.WithField("slugs", slugs)
	if len(slugs) == 0 {
		l.Error("Create")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	class := slugs[0]
	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		l.Error("Bad type: " + ct)
		w.WriteHeader(500)
		return
	}
	l = l.WithField("content length", r.ContentLength)
	if r.ContentLength == 0 {
		l.Error("Create")
		w.WriteHeader(500)
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.WithError(err).Error("Create")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, err := uuid.NewRandom()
	if err != nil {
		l.Error("Create")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	l = l.WithField("id", id)
	var data map[string]interface{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		l.WithField("json", string(b)).WithError(err).Error("Create")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = rest.store.Set(class, &store.Document{
		UID:  &id,
		Data: data,
	})
	if err != nil {
		l.WithError(err).Error("store.Set")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	l.Info("Create")
	fmt.Fprintf(w, `"%s"`, id.String())
}
