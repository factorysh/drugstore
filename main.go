package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/factorysh/drugstore/conf"
	"github.com/factorysh/drugstore/rest"
	"github.com/factorysh/drugstore/store"
	"github.com/factorysh/drugstore/version"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println(version.Version())
		return
	}
	cfgPath := os.Getenv("CONFIG")
	if cfgPath == "" {
		cfgPath = "/etc/drugstore.yml"
	}
	cfg, err := conf.Read(cfgPath)
	if err != nil {
		panic(err)
	}
	s, err := store.New(cfg.PostgresqlDSN)
	if err != nil {
		panic(err)
	}
	for k, v := range cfg.Classes {
		s.Class(k, v)
	}
	r, err := rest.New(s)
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", r.Handler())
	fmt.Println(cfg.Listen)
	http.ListenAndServe(cfg.Listen, nil)
}
