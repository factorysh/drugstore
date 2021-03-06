package store

import (
	"errors"

	"github.com/davecgh/go-spew/spew"
	"github.com/dop251/goja"
)

func (s *Store) ByMapReduce(class string, path []string, js string) ([]interface{}, error) {
	docs, err := s.GetByPath(class, path...)
	if err != nil {
		return nil, err
	}
	vm := goja.New()
	values := make([]interface{}, 0)
	emit := func(line interface{}) {
		values = append(values, line)
	}
	vm.Set("emit", emit)
	vm.Set("logs", func(stuff interface{}) {
		spew.Dump(stuff)
	})
	prog, err := goja.Compile("map.js", js, true)
	if err != nil {
		return nil, err
	}
	_, err = vm.RunProgram(prog)
	if err != nil {
		return nil, err
	}
	mapjs := vm.Get("map")
	_map, ok := goja.AssertFunction(mapjs)
	if !ok {
		return nil, errors.New("map must be a function")
	}

	for _, doc := range docs {
		_, err := _map(nil, vm.ToValue(doc.Data))
		if err != nil {
			return nil, err
		}
	}
	return values, nil

}
