package config

import (
	"log"
	"os"
	"regexp"

	"github.com/joho/godotenv"
)

var PROJECT_DIR_NAME = "black-flag"

// Solution: https://stackoverflow.com/a/68347834/2332999
func LoadEnvConfig() {
	projectName := regexp.MustCompile(`^(.*` + PROJECT_DIR_NAME + `)`)
	currentWorkDirectory, _ := os.Getwd()
	rootPath := projectName.Find([]byte(currentWorkDirectory))

	err := godotenv.Load(string(rootPath) + `/.env`)
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}
