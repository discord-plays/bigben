package bigben

import (
	"github.com/discord-plays/bigben/message"
)

func (b *BigBen) cronNewYears() {
	b.messageNotification("New Year's", message.SendNewYearNotification)()
}
