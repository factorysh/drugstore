package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/factorysh/drugstore/store"
	log "github.com/sirupsen/logrus"
)

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
	fmt.Fprint(w, "{}")
}
