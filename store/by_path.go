package store

import (
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// getSQLByPath build sql SELECT with a paths.
func (s *Store) getSQLByPath(class string, paths ...string) (string, error) {
	pathz, ok := s.paths[class]
	if !ok {
		return "", fmt.Errorf("Unknown class : %s", class)
	}
	if len(paths) > len(pathz) {
		return "", fmt.Errorf("Path too long : %s", paths)
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
	return buf.String(), nil
}

func (s *Store) GetByPath(class string, paths ...string) ([]Document, error) {
	l := log.WithField("class", class).WithField("paths", paths)
	sql, err := s.getSQLByPath(class, paths...)
	if err != nil {
		l.WithError(err).Error("GetByPath")
		return nil, err
	}
	l = l.WithField("sql", sql)
	var documents []RawDocument
	err = s.db.Select(&documents, sql)
	if err != nil {
		l.WithError(err).Error("GetByPath")
		return nil, err
	}
	l.WithField("#documents", len(documents)).Info("GetByPath")
	return raw2docs(documents)
}
