package store

import (
	"fmt"
	"strings"

	jmespath "github.com/jmespath/go-jmespath"
	log "github.com/sirupsen/logrus"
)

func (s *Store) GetByJMEspath(path string) ([]Document, error) {
	p := s.startPath.FindString(path)
	var docs []Document
	var err error
	l := log.WithField("path", path)
	if p != "" {
		if strings.HasSuffix(p, ".") {
			p = p[:len(p)-1]
		}
		pp := strings.Split(p, ".")
		docs, err = s.GetByPath(pp...)
		if err != nil {
			l.WithError(err).Error("GetByJMEspath")
			return nil, err
		}
	}
	data := make(map[string]interface{})
	for _, d := range docs {
		for _, p := range s.paths {
			k, ok := d.Data[p]
			if !ok {
				err := fmt.Errorf("Can't find key %s", p)
				l.WithError(err).Error("GetByJMEspath")
				return nil, err
			}
			fmt.Println(k)
			kk, ok := k.(string)
			if !ok {
				err := fmt.Errorf("Key is not a string : %s => %s", p, k)
				l.WithError(err).Error("GetByJMEspath")
				return nil, err
			}
			data[kk] = d
		}
	}
	jm, err := jmespath.Search(path, data)
	if err != nil {
		l.WithError(err).Error("GetByJMEspath")
		return nil, err
	}
	fmt.Println(jm)
	fmt.Println(data)
	return nil, nil
}
