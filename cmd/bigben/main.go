package main

import (
	"github.com/MrMelon54/BigBen/tables"
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
	err = engine.Sync(&tables.BongLog{}, &tables.GuildSettings{}, &tables.RoleLog{})
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

	ben, err := NewBigBen(engine, os.Getenv("TOKEN"), appId, guildId)
	if err != nil {
		log.Fatalln(err)
	}
	ben.RunAndBlock()
}
