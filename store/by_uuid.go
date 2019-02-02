package store

import (
	"bytes"

	"github.com/google/uuid"
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
	err := s.db.Select(&documents, buf.String())
	if err != nil {
		return nil, err
	}
	return raw2docs(documents)
}
