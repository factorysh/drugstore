package schema

import (
	"bufio"
	"bytes"
	"fmt"

	"gopkg.in/yaml.v2"
)

// Column has a name and can be a key
type Column struct {
	Type string `yaml:"type"`
	Key  bool   `yaml:"key"`
}

// Schema describes you stored stuff
type Schema struct {
	Name   string
	Values map[string]Column
}

// New Schema
func New(name string, raw []byte) (*Schema, error) {
	schema := Schema{Name: name}
	err := yaml.Unmarshal(raw, &schema.Values)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}

// DDL query
func (s Schema) DDL() (string, error) {
	buff := bytes.Buffer{}
	w := bufio.NewWriter(&buff)
	uniques := make([]string, 0)
	fmt.Fprintf(w, `CREATE TABLE IF NOT EXISTS %s (
		  id INT PRIMARY KEY`, s.Name)
	versions := make([]string, 0)
	for name, column := range s.Values {
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
		if column.Key {
			uniques = append(uniques, name)
		}
	}
	if len(uniques) > 0 {
		w.WriteString(`,
		UNIQUE(`)
		for i, unique := range uniques {
			fmt.Fprintf(w, `"%s"`, unique)
			if i < len(uniques)-1 {
				w.WriteString(", ")
			}
		}
		w.WriteString(")")
	}
	w.WriteString("\n);\n")
	for _, version := range versions {
		fmt.Fprintf(w, `
CREATE TABLE IF NOT EXISTS %s_%s (
  %s INT REFERENCES %s (id),
  version TEXT,
  name TEXT
);
  `, s.Name, version, s.Name, s.Name)
	}
	w.Flush()
	return buff.String(), nil
}

// Get query and arguments
func (s Schema) Get(doc map[string]interface{}) (string, []interface{}, error) {
	buff := bytes.Buffer{}
	w := bufio.NewWriter(&buff)
	fmt.Fprintf(w, `SELECT id FROM %s WHERE `, s.Name)
	keys := make([]interface{}, 0)
	for name, column := range s.Values {
		if column.Key {
			keys = append(keys, name)
		}
	}
	for i, key := range keys {
		fmt.Fprintf(w, `"%s"=$%d`, key, i+1)
		if i < len(keys)-1 {
			w.WriteString(" AND ")
		}
	}
	w.Flush()
	return buff.String(), keys, nil
}

// Set query and arguments
func (s Schema) Set(doc map[string]interface{}) (string, error) {
	values := make(map[string]interface{})
	for key, value := range doc {
		_, ok := s.Values[key]
		if !ok { // not in the schema
			continue
		}
		switch s.Values[key].Type {
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
		case "versions":
			v, ok := value.(map[string]interface{})
			if !ok {
				return "", fmt.Errorf("Not a versions : %p", value)
			}
			for k, vv := range v {
				_, ok := vv.(string)
				if !ok {
					return "", fmt.Errorf("Not a version : %s => %p", k, vv)
				}
			}
			values[key] = v
		}
	}
	buff := bytes.Buffer{}
	w := bufio.NewWriter(&buff)
	fmt.Fprintf(w, "INSERT INTO %s (", s.Name)
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
