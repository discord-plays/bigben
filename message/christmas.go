package message

import (
	_ "embed"
	"github.com/discord-plays/bigben/database"
	"github.com/discord-plays/bigben/logger"
	"github.com/disgoorg/disgo/bot"
	"sync"
	"time"
)

//go:embed christmas.txt
var christmasMessage string

// SendChristmasNotification makes a Christmas notification and send it in a message
func SendChristmasNotification(client bot.Client, wg *sync.WaitGroup, conf database.Guild, oldYear, newYear int) {
	defer wg.Done()
	builder := MakeMessageNotification("Merry Christmas", christmasMessage, "https://twemoji.maxcdn.com/v/latest/72x72/1f384.png", 0x5c9238, oldYear, newYear, time.Date(newYear, time.December, 25, 0, 0, 0, 0, time.UTC))
	_, err := client.Rest().CreateMessage(conf.BongChannelID, builder.Build())
	if err != nil {
		logger.Logger.Error("SendChristmasNotification", "id", conf.ID, "channel id", conf.BongChannelID, "err", err)
		return
	}
}
