package store

import (
	"fmt"
	"strings"

	"github.com/dolmen-go/jsonptr"
)

func (s *Store) GetByPtr(class string, paths ...string) (interface{}, error) {
	pathz, ok := s.paths[class]
	if !ok {
		return nil, fmt.Errorf("Unknown class : %s", class)
	}
	if len(paths) <= len(pathz) {
		return s.GetByPath(class, paths...)
	}
	docs, err := s.GetByPath(class, paths[:len(pathz)]...)
	if err != nil {
		return nil, err
	}
	return jsonptr.Get(docs, strings.Join(paths[len(pathz):], "/"))
}
