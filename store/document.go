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
