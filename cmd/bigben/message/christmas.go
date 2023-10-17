package message

import (
	_ "embed"
	"github.com/MrMelon54/bigben/tables"
	"github.com/disgoorg/disgo/bot"
	"log"
	"sync"
	"time"
)

//go:embed christmas.txt
var christmasMessage string

// SendChristmasNotification makes a Christmas notification and send it in a message
func SendChristmasNotification(client bot.Client, wg *sync.WaitGroup, conf tables.GuildSettings, oldYear, newYear int) {
	defer wg.Done()
	builder := MakeMessageNotification("Merry Christmas", christmasMessage, "https://twemoji.maxcdn.com/v/latest/72x72/1f384.png", 0x5c9238, oldYear, newYear, time.Date(newYear, time.December, 25, 0, 0, 0, 0, time.UTC))
	_, err := client.Rest().CreateMessage(conf.BongChannelId, builder.Build())
	if err != nil {
		log.Printf("[sendChristmasNotification(\"%s/%s\")] Error: %s\n", conf.GuildId, conf.BongChannelId, err)
		return
	}
}
