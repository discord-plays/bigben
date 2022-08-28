package main

import (
	"github.com/MrMelon54/BigBen/tables"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
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

	ben, err := NewBigBen(engine, os.Getenv("TOKEN"), os.Getenv("APP_ID"), os.Getenv("GUILD_ID"))
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("[Main] BigBen is now bonging. Press CTRL-C for maintenance.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	_ = ben.Exit()
}
