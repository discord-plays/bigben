package message

import (
	_ "embed"
	"github.com/discord-plays/bigben/database"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"log"
	"sync"
	"time"
)

//go:embed new-years.txt
var newYearsMessage string

// SendNewYearNotification makes a New Year notification and send it in a message
func SendNewYearNotification(client bot.Client, wg *sync.WaitGroup, conf database.Guild, oldYear, newYear int) {
	defer wg.Done()
	builder := MakeMessageNotification("Happy New Year's", newYearsMessage, "https://twemoji.maxcdn.com/v/latest/72x72/1f386.png", 0xd4af37, oldYear, newYear, time.Date(newYear, time.January, 1, 0, 0, 0, 0, time.UTC))
	a := []discord.InteractiveComponent{discord.NewLinkButton("Big Ben Archive", "https://bigben.mrmelon54.com").WithEmoji(discord.ComponentEmoji{Name: ""})}
	builder.SetContainerComponents(discord.NewActionRow(a...))
	_, err := client.Rest().CreateMessage(conf.BongChannelID, builder.Build())
	if err != nil {
		log.Printf("[SendNewYearNotification(\"%s/%s\")] Error: %s\n", conf.ID, conf.BongChannelID, err)
		return
	}
}
