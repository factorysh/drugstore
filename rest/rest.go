package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"

	_ "github.com/factorysh/drugstore/statik"
	"github.com/factorysh/drugstore/store"
	_fs "github.com/rakyll/statik/fs"
	log "github.com/sirupsen/logrus"
)

type REST struct {
	store *store.Store
	mux   *http.ServeMux
}

func New(store *store.Store) (*REST, error) {
	r := &REST{
		store: store,
		mux:   http.NewServeMux(),
	}
	var (
		fs  http.FileSystem
		err error
	)
	public := os.Getenv("PUBLIC")
	if public == "" {
		fs, err = _fs.New()
		if err != nil {
			return nil, err
		}
	} else {
		fmt.Println("Using local folder: ", public)
		fs = http.Dir(public)
	}
	r.mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(fs)))
	r.mux.HandleFunc("/", r.Main)
	return r, nil
}

func (rest *REST) Handler() func(w http.ResponseWriter, r *http.Request) {
	return rest.mux.ServeHTTP
}

// Handler routes all handlers
func (rest *REST) Main(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		home(w)
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

func home(w http.ResponseWriter) {
	// http://anime.en.utf8art.com/arc/ghibli_6.html
	w.Write([]byte(
		"	　　　　　　　　　　　　　　　 ﾍ\n" +
			"　　　　　　　　　　　　　　 ﾍ　　　/　|\n" +
			"　　　　　　　　　　　　　 / ｜　 /　　|\n" +
			"　　　　　　　　　 }YL　 ﾉ　　|　ﾉ 　 　|\n" +
			"　　　　　　　　　ﾉ　　ヽﾐ}　F′〉　 ｯ┘\n" +
			"　 　 　　　　　　{^^ . -┴┴‐ミ　　ﾐ.._\n" +
			"　 　 　　　　　　> ´　　　　　　　　　　ミ､\n" +
			"　　　　　　　　/　　　　　　　　　　　　　 ﾐ､\n" +
			"　　　　　　　 ﾉ　　p￣ヽ_　　　　　　　　　ﾐ､\n" +
			"　　　　　rﾍ⌒　　 `ー ′　　　　　　　　　 ﾐ､\n" +
			"　　　　ﾆ{^　　　　　　　　　　　　　　　　　　 ﾐ､\n" +
			"　　　　 〈､_　　　＝三二_ー--　　　　　　　　 l\n" +
			"　　　　∠_　　　　　ｰ＝= 二_ｰ\n" +
			"　 　／　　 ¨ヾ､\n" +
			"　 ﾉﾍ　　　　　　ヽ\n"))
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
