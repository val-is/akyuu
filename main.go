package main

import (
	"log"

	"github.com/go-martini/martini"
	akyuu "github.com/val-is/akyuu/src"
)

func main() {
	fsClient, err := akyuu.NewFsClient("akyuu/fsconfig.json")
	if err != nil {
		log.Fatalf("Error loading FS client: %s", err)
	}

	m := martini.Classic()
	akyuu.BuildRoutes(m, &fsClient)
	m.Run()
}
