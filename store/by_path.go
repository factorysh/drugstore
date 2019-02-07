package store

import (
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"
)

func (s *Store) GetByPath(class string, paths ...string) ([]Document, error) {
	pathz, ok := s.paths[class]
	if !ok {
		return nil, fmt.Errorf("Unknown class : %s", class)
	}
	if len(paths) > len(pathz) {
		return nil, fmt.Errorf("Path too long : %s", paths)
	}
	buf := bytes.NewBuffer([]byte(`
		SELECT *
		FROM document
	`))
	if len(paths) > 0 {
		size := 0
		for i, p := range paths {
			if p == "" {
				continue
			}
			// TODO assert p [a-zA-Z0-9:_]
			if size == 0 {
				buf.WriteString("WHERE ")
			}
			if size > 0 {
				buf.WriteString(" AND ")
			}
			buf.WriteString(` data @> '{"`)
			buf.WriteString(pathz[i])
			buf.WriteString(`": "`)
			buf.WriteString(p)
			buf.WriteString(`"}'`)
			size++
		}
	}
	l := log.WithField("sql", buf.String())
	var documents []RawDocument
	err := s.db.Select(&documents, buf.String())
	if err != nil {
		l.WithError(err).Error("GetByPath")
		return nil, err
	}
	l.Info("GetByPath")
	return raw2docs(documents)
}
