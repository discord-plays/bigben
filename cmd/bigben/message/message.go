package message

import (
	"github.com/disgoorg/disgo/discord"
	"strings"
	"text/template"
	"time"
)

// MakeMessageNotification makes a generic notification with the old year and new year values
func MakeMessageNotification(title, message, thumbnail string, colour, oldYear, newYear int, timestamp time.Time) *discord.MessageCreateBuilder {
	tmpl := template.New("description")
	_, err := tmpl.Parse(message)
	if err != nil {
		return nil
	}

	b := new(strings.Builder)
	err = tmpl.Execute(b, struct {
		OldYear int
		NewYear int
	}{oldYear, newYear})
	if err != nil {
		return nil
	}

	embed := discord.NewEmbedBuilder()
	embed.SetTitle(title)
	embed.SetAuthor("Melon", "https://mrmelon54.com", "https://cdn.discordapp.com/avatars/222344019458392065/ddc5b5cb27f8b0d1df7521b192940427.png?size=256")
	embed.SetDescriptionf(b.String())
	embed.SetColor(colour)
	embed.SetTimestamp(timestamp)
	embed.SetThumbnail(thumbnail)
	builder := discord.NewMessageCreateBuilder()
	builder.SetEmbeds(embed.Build())
	return builder
}
