package bigben

import (
	"context"
	"github.com/discord-plays/bigben/database"
	"github.com/discord-plays/bigben/logger"
	"github.com/disgoorg/disgo/bot"
	"sync"
	"time"
)

type messageNotificationCallback func(client bot.Client, wg *sync.WaitGroup, conf database.Guild, oldYear int, newYear int)

func (b *BigBen) messageNotification(name string, call messageNotificationCallback) func() {
	return func() {
		logger.Logger.Info("Sending " + name + " Notification")
		now := time.Now()
		year := now.Year()
		all, err := b.engine.GetAllGuilds(context.Background())
		if err != nil {
			logger.Logger.Error("GetAllGuilds", "err", err)
			return
		}
		wg := &sync.WaitGroup{}
		for _, i := range all {
			wg.Add(1)
			go call(b.client, wg, i, year-1, year)
		}
		wg.Wait()
	}
}
