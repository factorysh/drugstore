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

func (s Schema) Set(doc map[string]interface{}) error {
	for key, value := range doc {
		_, ok := s[key]
		if !ok { // not in the schema
			continue
		}
		switch s[key].Type {
		case "string":
			v, ok := value.(int64)
			if !ok {
				return fmt.Errorf("Not an int : %p", value)
			}
			fmt.Println(v)

		}
	}
	return nil
}
