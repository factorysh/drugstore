package store

import (
	"bytes"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func (s *Store) GetByUUID(uuids ...uuid.UUID) ([]Document, error) {
	var documents []RawDocument
	buf := bytes.NewBuffer([]byte(`
		SELECT *
		FROM document
		WHERE uid IN (
	`))
	size := len(uuids)
	for i, uid := range uuids {
		buf.WriteRune('\'')
		buf.WriteString(uid.String())
		buf.WriteRune('\'')
		if i+1 < size {
			buf.WriteRune(',')

		}
	}
	buf.WriteRune(')')
	l := log.WithField("sql", buf.String())
	err := s.db.Select(&documents, buf.String())
	if err != nil {
		l.WithError(err).Error("GetByUUID")
		return nil, err
	}
	l.Info("GetByUUID")
	return raw2docs(documents)
}
