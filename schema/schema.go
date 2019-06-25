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
	buff.WriteString("CREATE TABLE ")
	buff.WriteString(table)
	buff.WriteString(" (\n  id integer PRIMARY KEY")
	versions := make([]string, 0)
	for name, column := range s {
		if column.Type == "versions" {
			versions = append(versions, name)
			continue
		}
		buff.WriteString(",\n  ")
		buff.WriteString(name)
		buff.WriteRune(' ')
		switch column.Type {
		case "string":
			buff.WriteString("text")
		case "boolean":
			buff.WriteString("boolean")
		case "integer":
			buff.WriteString("integer")
		}
	}
	buff.WriteString("\n);\n")
	w := bufio.NewWriter(&buff)
	for _, version := range versions {
		fmt.Fprintf(w, `
CREATE TABLE %s_%s (
  %s integer REFERENCES %s (id),
  version string,
  name string
);
  `, table, version, table, table)
	}
	w.Flush()
	return buff.String(), nil
}
