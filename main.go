package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/go-martini/martini"
	"github.com/spf13/viper"
	akyuu "github.com/val-is/akyuu/src"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config: %s", err)
	}

	m := martini.Classic()

	fsListingPath := viper.GetString("fsdb.listing_path")
	storageDir := viper.GetString("fsdb.storage_dir")
	for i := 0; i < akyuu.NFileTypes; i++ {
		createDir(fmt.Sprintf("%s/%d", storageDir, i))
	}
	fsClient, err := akyuu.NewFsClient(fsListingPath, storageDir, fileExists(fsListingPath))
	if err != nil {
		log.Fatalf("Error loading FS client: %s", err)
	}
	m.Map(&fsClient)

	tokenPath := viper.GetString("tokenreg.token_path")
	tokenReg, err := akyuu.NewTokenReg(tokenPath, fileExists(tokenPath))
	if err != nil {
		log.Fatalf("Error loading token reg: %s", err)
	}
	m.Map(&tokenReg)

	akyuu.BuildRoutes(m)

	m.Run()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else if err != nil {
		log.Fatalf("Error statting file %s: %s", path, err)
	}
	return false
}

func createDir(path string) {
	err := os.MkdirAll(path, 0644)
	if err != nil {
		log.Fatalf("Error creating dir %s: %s", path, err)
	}
}
