package schema

import (
	"bufio"
	"bytes"
	"fmt"

	"gopkg.in/yaml.v2"
)

type Column struct {
	Type string `yaml:"type"`
	Key  bool   `yaml:"key"`
}

type Schema struct {
	name string
	values map[string]Column
}

func New(name string, raw []byte) (*Schema, error) {
	schema := Schema{ name: name}
	err := yaml.Unmarshal(raw, &schema.values)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}

func (s Schema) DDL() (string, error) {
	buff := bytes.Buffer{}
	w := bufio.NewWriter(&buff)
	fmt.Fprintf(w, `CREATE TABLE IF NOT EXISTS %s (
		  id INT PRIMARY KEY`, s.name)
	versions := make([]string, 0)
	for name, column := range s.values {
		if column.Type == "versions" {
			versions = append(versions, name)
			continue
		}
		fmt.Fprintf(w, `,
		"%s" `, name)
		switch column.Type {
		case "string":
			w.WriteString("TEXT")
		case "boolean":
			w.WriteString("BOOLEAN")
		case "integer":
			w.WriteString("INTEGER")
		}
	}
	w.WriteString("\n);\n")
	for _, version := range versions {
		fmt.Fprintf(w, `
CREATE TABLE IF NOT EXISTS %s_%s (
  %s INT REFERENCES %s (id),
  version TEXT,
  name TEXT
);
  `, s.name, version, s.name, s.name)
	}
	w.Flush()
	return buff.String(), nil
}

func (s Schema) Set(doc map[string]interface{}) (string, error) {
	values := make(map[string]interface{})
	for key, value := range doc {
		_, ok := s.values[key]
		if !ok { // not in the schema
			continue
		}
		switch s.values[key].Type {
		case "integer":
			v, ok := value.(int64)
			if !ok {
				return "", fmt.Errorf("Not an int : %p", value)
			}
			values[key] = v
		case "string":
			v, ok := value.(string)
			if !ok {
				return "", fmt.Errorf("Not a string : %p", value)
			}
			values[key] = v
		case "boolean":
			v, ok := value.(bool)
			if !ok {
				return "", fmt.Errorf("Not a boolean : %p", value)
			}
			values[key] = v
		}
	}
	buff := bytes.Buffer{}
	w := bufio.NewWriter(&buff)
	fmt.Fprintf(w, "INSERT INTO %s (", s.name)
	cpt := len(values)
	for k := range values {
		w.WriteString(k)
		cpt--
		if cpt > 0 {
			w.WriteString(", ")
		}
	}
	w.WriteString(") VALUES (")
	cpt = len(values)
	for k := range values {
		w.WriteString(":")
		w.WriteString(k)
		cpt--
		if cpt > 0 {
			w.WriteString(", ")
		}
	}
	w.WriteString(") ON CONFLICT DO UPDATE SET ")
	cpt = len(values)
	for k := range values {
		fmt.Fprintf(w, " %s = :%s", k, k)
		cpt--
		if cpt > 0 {
			w.WriteString(", ")
		}
	}
	w.WriteString(";")
	w.Flush()
	return buff.String(), nil

}
