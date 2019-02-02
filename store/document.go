package store

import (
	"encoding/json"

	"github.com/google/uuid"
)

type RawDocument struct {
	UID  uuid.UUID       `db:"uid"`
	Data json.RawMessage `db:"data"`
}

type Document struct {
	UID  uuid.UUID
	Data map[string]interface{}
}

func raw2docs(raws []RawDocument) ([]Document, error) {
	docs := make([]Document, len(raws))
	for i, d := range raws {
		var dd map[string]interface{}
		err := json.Unmarshal(d.Data, &dd)
		if err != nil {
			return nil, err
		}
		docs[i] = Document{
			UID:  d.UID,
			Data: dd,
		}
	}
	return docs, nil
}
