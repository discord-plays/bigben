package main

import (
	"github.com/charmbracelet/log"
	"github.com/discord-plays/bigben"
	"github.com/discord-plays/bigben/logger"
	"github.com/disgoorg/snowflake/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		logger.Logger.Fatal("Loading .env", "err", err)
	}

	if os.Getenv("DEBUG_MODE") == "1" {
		logger.Logger.Debug("Activating DEBUG mode")
		logger.Logger.SetLevel(log.DebugLevel)
	}
	logger.Logger.Info("Loading database")
	dbEnv := os.Getenv("DB")
	db, err := bigben.InitDB(dbEnv)
	if err != nil {
		logger.Logger.Fatal("Failed to load database", "err", err)
	}

	appId, err := snowflake.Parse(os.Getenv("APP_ID"))
	if err != nil {
		logger.Logger.Fatal("Invalid APP_ID", "err", err)
	}
	guildId, err := snowflake.Parse(os.Getenv("GUILD_ID"))
	if err != nil {
		guildId = 0
	}

	ben, err := bigben.NewBigBen(db, os.Getenv("TOKEN"), os.Getenv("UPLOAD_TOKEN"), os.Getenv("STATUS_PUSH"), appId, guildId)
	if err != nil {
		logger.Logger.Fatal("Failed to start", "err", err)
	}
	ben.RunAndBlock()
}
