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

type Schema map[string]Column

func New(raw []byte) (*Schema, error) {
	var schema Schema
	err := yaml.Unmarshal(raw, &schema)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}

func (s Schema) DDL(table string) (string, error) {
	buff := bytes.Buffer{}
	buff.WriteString("CREATE TABLE IF NOT EXISTS ")
	buff.WriteString(table)
	buff.WriteString(" (\n  id INT PRIMARY KEY")
	versions := make([]string, 0)
	for name, column := range s {
		if column.Type == "versions" {
			versions = append(versions, name)
			continue
		}
		buff.WriteString(",\n  \"")
		buff.WriteString(name)
		buff.WriteString(`" `)
		switch column.Type {
		case "string":
			buff.WriteString("TEXT")
		case "boolean":
			buff.WriteString("BOOLEAN")
		case "integer":
			buff.WriteString("INTEGER")
		}
	}
	buff.WriteString("\n);\n")
	w := bufio.NewWriter(&buff)
	for _, version := range versions {
		fmt.Fprintf(w, `
CREATE TABLE IF NOT EXISTS %s_%s (
  %s INT REFERENCES %s (id),
  version TEXT,
  name TEXT
);
  `, table, version, table, table)
	}
	w.Flush()
	return buff.String(), nil
}

func (s Schema) Set(doc map[string]interface{}) (string, error) {
	values := make(map[string]interface{})
	for key, value := range doc {
		_, ok := s[key]
		if !ok { // not in the schema
			continue
		}
		switch s[key].Type {
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
	fmt.Fprintf(w, "INSERT INTO %s (", "toto")
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
