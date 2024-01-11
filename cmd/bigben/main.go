package main

import (
	"fmt"
	"github.com/discord-plays/bigben/tables"
	"github.com/disgoorg/snowflake/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"strings"
	"xorm.io/xorm"
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
	var engine *xorm.Engine
	if strings.HasPrefix(dbEnv, "sqlite:") {
		engine, err = xorm.NewEngine("sqlite3", strings.TrimPrefix(dbEnv, "sqlite:"))
	} else if strings.HasPrefix(dbEnv, "mysql:") {
		engine, err = xorm.NewEngine("mysql", strings.TrimPrefix(dbEnv, "mysql:"))
	} else {
		log.Fatalln("[Main] Only mysql and sqlite are supported")
	}
	if err != nil {
		log.Fatalf("Unable to load database (\"%s\"): %s\n", dbEnv, err)
	}
	err = engine.Sync(&tables.BongLog{}, &tables.GuildSettings{}, &tables.LeaderboardUploads{}, &tables.RoleLog{}, &tables.UserLog{})
	if err != nil {
		log.Fatalf("Unable to sync database: %s\n", err)
	}

	appId, err := snowflake.Parse(os.Getenv("APP_ID"))
	if err != nil {
		log.Fatalf("Unable to parse APP_ID: %s\n", err)
	}
	guildId, err := snowflake.Parse(os.Getenv("GUILD_ID"))
	if err != nil {
		guildId = 0
	}

	ben, err := NewBigBen(engine, os.Getenv("TOKEN"), os.Getenv("UPLOAD_TOKEN"), os.Getenv("STATUS_PUSH"), appId, guildId)
	if err != nil {
		log.Fatalln(err)
	}
	ensureBackupsAreUploaded(ben)
	ben.RunAndBlock()
}

func ensureBackupsAreUploaded(ben *BigBen) {
	toUploadYears := make([]int, 0)
	err := ben.engine.Iterate(&tables.LeaderboardUploads{}, func(idx int, bean interface{}) error {
		row, ok := bean.(*tables.LeaderboardUploads)
		if !ok {
			return fmt.Errorf("failed to convert to iterating type")
		}
		if !*row.Sent {
			toUploadYears = append(toUploadYears, row.Year)
		}
		return nil
	})
	if err != nil {
		log.Printf("[ensureBackupsAreUploaded()] Failed to iterate over leaderboard years: %s\n", err)
		return
	}
	for _, i := range toUploadYears {
		generateAndUploadBackup(ben.engine, i, ben.uploadToken)
	}
}
