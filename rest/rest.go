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
)

type REST struct {
	store *store.Store
}

func (rest *REST) GetByPath(w http.ResponseWriter, r *http.Request) {
	slugs := strings.Split(r.RequestURI, "/")
	if len(slugs) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	class := slugs[0]
	slugs = slugs[1:]
	docs, err := rest.store.GetByPath(class, slugs...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf := bytes.NewBuffer([]byte("[\n"))
	for i, doc := range docs {
		j, err := json.Marshal(doc.Data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		buf.Write(j)
		if i < len(docs) {
			buf.WriteString(",\n")
		}
	}
	buf.WriteString("\n]")
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes())
}

func (rest *REST) Query(w http.ResponseWriter, r *http.Request) {
	slugs := strings.Split(r.RequestURI, "/")
	if len(slugs) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	class := slugs[0]
	q := r.URL.Query().Get("q")
	resp, err := rest.store.GetByJMEspath(class, q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	j, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func (rest *REST) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	slugs := strings.Split(r.RequestURI, "/")
	if len(slugs) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	class := slugs[0]
	fmt.Println(r.Header.Get("content-type"))
	fmt.Println("r", r)
	fmt.Println("length", r.ContentLength)
	body, err := r.GetBody()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b, err := ioutil.ReadAll(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, err := uuid.NewRandom()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var data map[string]interface{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rest.store.Set(class, &store.Document{
		UID:  id,
		Data: data,
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, `"%s"`, id.String())
}
