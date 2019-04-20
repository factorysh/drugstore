package main

import (
	"fmt"
	"os"

	"github.com/factorysh/drugstore/version"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println(version.Version())
	}
}
