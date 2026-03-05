package util

import (
	"log"
	"os"
)

func RequireEnvs() {
	envs := []string{
		"DB_URL",
		"PORT",
	}

	for _, env := range envs {
		if os.Getenv(env) == "" {
			log.Fatal("Missing required env:", env)
		}
	}
}