package store

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
)

func (s *Store) GetByUUID(uuids ...uuid.UUID) ([]Document, error) {
	var documents []RawDocument
	ids := make([]string, len(uuids))
	for i, uid := range uuids {
		// FIXME: OMG it's ugly!
		ids[i] = "'" + uid.String() + "'"
	}
	err := s.db.Select(&documents, `
		SELECT *
		FROM document
		WHERE uid IN (`+strings.Join(ids, ",")+")")
	if err != nil {
		return nil, err
	}
	// FIXME it's ugly, sqlx job is done twice
	ds := make([]Document, len(documents))
	for i, d := range documents {
		var dd map[string]interface{}
		err = json.Unmarshal(d.Data, &dd)
		if err != nil {
			return nil, err
		}
		ds[i] = Document{
			UID:  d.UID,
			Data: dd,
		}
	}
	return ds, nil
}
