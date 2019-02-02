package store

import (
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"
)

func (s *Store) GetByPath(paths ...string) ([]Document, error) {
	if len(paths) > len(s.paths) {
		return nil, fmt.Errorf("Path too long : %s", paths)
	}
	buf := bytes.NewBuffer([]byte(`
		SELECT *
		FROM document
	`))
	if len(paths) > 0 {
		buf.WriteString("WHERE ")

		for i, p := range paths {
			buf.WriteString(` data @> '{"`)
			buf.WriteString(s.paths[i])
			buf.WriteString(`": "`)
			buf.WriteString(p)
			buf.WriteString(`"}'`)
			if i+1 < len(paths) {
				buf.WriteString(" AND ")
			}
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
