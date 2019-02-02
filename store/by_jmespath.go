package store

import (
	"fmt"
	"strings"

	jmespath "github.com/jmespath/go-jmespath"
)

func (s *Store) GetByJMEspath(path string) ([]Document, error) {
	/*
		n, err := s.parser.Parse(path)
		if err != nil {
			return nil, err
		}
		fmt.Println(n)
	*/
	p := s.startPath.FindString(path)
	var docs []Document
	var err error
	if p != "" {
		if strings.HasSuffix(p, ".") {
			p = p[:len(p)-1]
		}
		pp := strings.Split(p, ".")
		docs, err = s.GetByPath(pp...)
		if err != nil {
			return nil, err
		}
	}
	data := make(map[string]interface{})
	for _, d := range docs {
		for _, p := range s.paths {
			k, ok := d.Data[p]
			if !ok {
				return nil, fmt.Errorf("Can't find key %s", p)
			}
			fmt.Println(k)
			kk, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("Key is not a string : %s => %s", p, k)
			}
			data[kk] = d
		}
	}
	jm, err := jmespath.Search(path, data)
	if err != nil {
		return nil, err
	}
	fmt.Println(jm)
	fmt.Println(data)
	return nil, nil
}
