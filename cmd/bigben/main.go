package main

import (
	"github.com/discord-plays/bigben"
	"github.com/disgoorg/snowflake/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}

	if os.Getenv("DEBUG_MODE") == "1" {
		log.Println("[Main] Activating DEBUG mode")
	}
	log.Println("[Main] Loading database")
	dbEnv := os.Getenv("DB")
	db, err := bigben.InitDB(dbEnv)
	if err != nil {
		log.Fatalln("[DatabaseError] ", err)
	}

	appId, err := snowflake.Parse(os.Getenv("APP_ID"))
	if err != nil {
		log.Fatalf("Unable to parse APP_ID: %s\n", err)
	}
	guildId, err := snowflake.Parse(os.Getenv("GUILD_ID"))
	if err != nil {
		guildId = 0
	}

	ben, err := bigben.NewBigBen(db, os.Getenv("TOKEN"), os.Getenv("UPLOAD_TOKEN"), os.Getenv("STATUS_PUSH"), appId, guildId)
	if err != nil {
		log.Fatalln(err)
	}
	ben.RunAndBlock()
}
