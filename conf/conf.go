package conf

import (
	"errors"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Conf struct {
	Listen        string              `yaml:"listen"`
	PostgresqlDSN string              `yaml:"postgresql_dsn"`
	Classes       map[string][]string `yaml:"classes"`
}

func New(data []byte) (*Conf, error) {
	var conf Conf
	err := yaml.Unmarshal(data, &conf)
	if err != nil {
		return nil, err
	}
	if conf.Listen == "" {
		conf.Listen = "127.0.0.1:5000"
	}
	if conf.PostgresqlDSN == "" {
		return nil, errors.New("postgresql_dsn is mandatory")
	}
	if len(conf.Classes) == 0 {
		return nil, errors.New("You need at least one class")
	}
	return &conf, nil
}

func Read(path string) (*Conf, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return New(data)
}
