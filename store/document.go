package store

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type RawDocument struct {
	UID   *uuid.UUID      `db:"uid"`
	Data  json.RawMessage `db:"data"`
	Mtime time.Time       `db:"mtime"`
	Ctime time.Time       `db:"ctime"`
	Class string          `db:"class"`
}

type Document struct {
	UID   *uuid.UUID
	Class string
	Data  map[string]interface{}
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
			UID:   d.UID,
			Class: d.Class,
			Data:  dd,
		}
	}
	return docs, nil
}
