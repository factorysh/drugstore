package store

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func (s *Store) GetByGJson(path string) (interface{}, error) {
	docs, err := s.guessByPath(path)
	l := log.WithField("path", path)
	if err != nil {
		l.WithError(err).Error("GetByGJson")
		return nil, err
	}
	data, err := s.Documents2tree(docs)
	if err != nil {
		l.WithError(err).Error("GetByGJson")
		return nil, err
	}
	json, err := json.Marshal(data)
	if err != nil {
		l.WithError(err).Error("GetByGJson")
		return nil, err
	}
	return gjson.GetBytes(json, path), nil
}
