package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/go-martini/martini"
	akyuu "github.com/val-is/akyuu/src"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	m := martini.Classic()

	fsClient, err := akyuu.NewFsClient("akyuu/fsconfig.json")
	if err != nil {
		log.Fatalf("Error loading FS client: %s", err)
	}
	m.Map(&fsClient)

	tokenReg, err := akyuu.NewTokenReg("akyuu/tokenconfig.json")
	if err != nil {
		log.Fatalf("Error loading token reg: %s", err)
	}
	m.Map(&tokenReg)

	akyuu.BuildRoutes(m)

	m.Run()
}
