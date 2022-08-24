package main

import (
	"fmt"
	"github.com/MrMelon54/BigBen"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}
	ben, err := bigben.NewBigBen(os.Getenv("TOKEN"), os.Getenv("APP_ID"), os.Getenv("GUILD_ID"))
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("BigBen is now bonging. Press CTRL-C for maintenance.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	_ = ben.Exit()
}
