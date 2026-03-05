package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/trnahnh/katana-id/internal/db"
	"github.com/trnahnh/katana-id/util"
)

func main() {
	godotenv.Load()
	util.RequireEnvs()

	ctx := context.Background()
	_, pool, err := db.Connect(ctx, os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	
}