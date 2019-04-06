package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/rpc"

	_patch "github.com/evanphx/json-patch"
	"github.com/factorysh/drugstore/store"
	"github.com/google/uuid"
)

type Service struct {
	store *store.Store
}

// New service
func New(s *store.Store) *Service {
	return &Service{
		store: s,
	}
}

// NewServer jsonrpc2
func NewServer(s *Service) *rpc.Server {
	server := rpc.NewServer()
	server.RegisterName("drugstore", s)
	return server
}

// Create a document, return its id
func (s *Service) Create(class string, document map[string]interface{}) (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	d := &store.Document{
		UID:  &u,
		Data: document,
	}
	err = s.store.Set(class, d)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// Update a document, with its id and values
// FIXME using a hash or a version for collision?
func (s *Service) Update(id string, document map[string]interface{}) error {
	u, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	d, err := s.store.GetByUUID(u)
	if err != nil {
		return err
	}
	if len(d) == 0 {
		return fmt.Errorf("Unknown document : %s", id)
	}
	s.store.Set(d[0].Class, &store.Document{
		UID:  &u,
		Data: document,
	})
	return nil
}

// Delete this document
func (s *Service) Delete(id string) error {
	u, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return s.store.Delete(u)
}

func (s *Service) Get(id string) (map[string]interface{}, error) {
	u, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	d, err := s.store.GetByUUID(u)
	if err != nil {
		return nil, err
	}
	return d[0].Data, nil
}

func (s *Service) Patch(id string, patch interface{}) error {
	rawPatch, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	u, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	d, err := s.store.GetByUUID(u)
	if err != nil {
		return err
	}
	rawDoc, err := json.Marshal(d[0].Data)
	if err != nil {
		return err
	}
	rawMerge, err := _patch.MergePatch(rawDoc, rawPatch)
	if err != nil {
		return err
	}
	var merge map[string]interface{}
	err = json.Unmarshal(rawMerge, &merge)
	if err != nil {
		return err
	}
	s.store.Set(d[0].Class, &store.Document{
		UID:  &u,
		Data: merge,
	})
	return nil

}
