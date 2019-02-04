package store

import (
	"fmt"
	"strings"

	jmespath "github.com/jmespath/go-jmespath"
	log "github.com/sirupsen/logrus"
)

func (s *Store) guessByPath(path string) ([]Document, error) {
	p := s.startPath.FindString(path)
	if p != "" {
		if strings.HasSuffix(p, ".") {
			p = p[:len(p)-1]
		}
		pp := strings.Split(p, ".")
		return s.GetByPath(pp...)
	} else {
		return s.GetByPath()
	}
}

func (s *Store) GetByJMEspath(path string) (interface{}, error) {
	docs, err := s.guessByPath(path)
	l := log.WithField("path", path)
	if err != nil {
		l.WithError(err).Error("GetByJMEspath")
		return nil, err
	}
	fmt.Println("docs", docs)
	data, err := s.Documents2tree(docs)
	if err != nil {
		l.WithError(err).Error("GetByJMEspath")
		return nil, err
	}
	jm, err := jmespath.Search(path, data)
	if err != nil {
		l.WithError(err).Error("GetByJMEspath")
		return nil, err
	}
	return jm, nil
}
