package conf

import (
	"gopkg.in/yaml.v2"
)

type Conf struct {
	PostgresqlDSN string              `yaml:"postgresql_dsn"`
	Classes       map[string][]string `yaml:"classes"`
}

func New(data []byte) (*Conf, error) {
	var conf Conf
	err := yaml.Unmarshal(data, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}
