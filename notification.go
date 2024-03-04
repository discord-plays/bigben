package bigben

import (
	"context"
	"github.com/discord-plays/bigben/database"
	"github.com/disgoorg/disgo/bot"
	"log"
	"sync"
	"time"
)

type messageNotificationCallback func(client bot.Client, wg *sync.WaitGroup, conf database.Guild, oldYear int, newYear int)

func (b *BigBen) messageNotification(name string, call messageNotificationCallback) func() {
	return func() {
		log.Printf("[messageNotification()] Sending %s Notification\n", name)
		now := time.Now()
		year := now.Year()
		all, err := b.engine.GetAllGuilds(context.Background())
		if err != nil {
			log.Printf("[messageNotification()] Error: %s\n", err)
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
