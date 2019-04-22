package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	_ "github.com/factorysh/drugstore/statik"
	"github.com/factorysh/drugstore/store"
	_fs "github.com/rakyll/statik/fs"
	log "github.com/sirupsen/logrus"
)

type REST struct {
	store *store.Store
	mux   *http.ServeMux
	fs    http.FileSystem
}

func New(store *store.Store) (*REST, error) {
	r := &REST{
		store: store,
		mux:   http.NewServeMux(),
	}
	var err error
	public := os.Getenv("PUBLIC")
	if public == "" {
		r.fs, err = _fs.New()
		if err != nil {
			return nil, err
		}
	} else {
		fmt.Println("Using local folder: ", public)
		r.fs = http.Dir(public)
	}
	r.mux.Handle("/_public/", http.StripPrefix("/_public/", http.FileServer(r.fs)))
	r.mux.HandleFunc("/_classes", r.classes)
	r.mux.HandleFunc("/", r.Main)
	return r, nil
}

func (rest *REST) Handler() func(w http.ResponseWriter, r *http.Request) {
	return rest.mux.ServeHTTP
}

func (rest *REST) classes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	j, err := json.Marshal(rest.store.Classes())
	if err != nil {
		w.WriteHeader(500)
	}
	w.Write(j)
}

// Handler routes all handlers
func (rest *REST) Main(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.FileServer(rest.fs).ServeHTTP(w, r)
		return
	}
	slugs := strings.Split(r.URL.Path, "/")[1:]
	class := slugs[0]
	if !rest.store.HasClass(class) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if r.Method == "GET" {
		if len(slugs) == 2 && slugs[1] == "_search" {
			rest.query(w, r)
			return
		}
		rest.getByPath(w, r)
		return
	}
	if r.Method == "POST" {
		rest.create(w, r)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

// GetByPath get an object, from its path
func (rest *REST) getByPath(w http.ResponseWriter, r *http.Request) {
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

// Query  GET /{class}/_search?q={query}
func (rest *REST) query(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	slugs := strings.Split(r.URL.Path, "/")
	l := log.WithField("url", r.URL.String()).WithField("slugs", slugs)
	if len(slugs) < 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	class := slugs[1]
	if slugs[2] != "_search" {
		w.WriteHeader(400)
		return
	}
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

func (rest *REST) create(w http.ResponseWriter, r *http.Request) {
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
	var data map[string]interface{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		l.WithField("json", string(b)).WithError(err).Error("Create")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = rest.store.Set(class, &store.Document{
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
	w.Write([]byte("{}"))
}
