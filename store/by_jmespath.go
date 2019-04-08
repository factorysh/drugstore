package store

import (
	"strings"

	jmespath "github.com/jmespath/go-jmespath"
	log "github.com/sirupsen/logrus"
)

func (s *Store) guessByPath(class string, path string) ([]Document, error) {
	p := s.startPath.FindString(path)
	if p != "" {
		if strings.HasSuffix(p, ".") {
			p = p[:len(p)-1]
		}
		pp := strings.Split(p, ".")
		return s.GetByPath(class, pp...)
	} else {
		return s.GetByPath(class)
	}
}

func (s *Store) GetByJMEspath(class string, path string) (interface{}, error) {
	l := log.WithField("class", class).WithField("path", path)
	docs, err := s.guessByPath(class, path)
	if err != nil {
		l.WithError(err).Error("GetByJMEspath")
		return nil, err
	}
	l = l.WithField("docs", docs)
	data, err := s.Documents2tree(class, docs)
	if err != nil {
		l.WithError(err).Error("GetByJMEspath")
		return nil, err
	}
	l = l.WithField("data", data)
	jm, err := jmespath.Search(path, data)
	if err != nil {
		l.WithError(err).Error("GetByJMEspath")
		return nil, err
	}
	l.WithField("jm", jm).Info("GetByJMEspath")
	return jm, nil
}
