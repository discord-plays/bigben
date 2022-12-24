package message

import (
	"github.com/disgoorg/disgo/discord"
	"strings"
	"text/template"
	"time"
)

func makeMessageNotification(title, message, thumbnail string, colour, oldYear, newYear int, timestamp time.Time) *discord.MessageCreateBuilder {
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
	embed.SetAuthor("Melon", "https://mrmelon54.com", "https://cdn.discordapp.com/avatars/222344019458392065/634a1f1256880daba803abb9330b76f4.png?size=256")
	embed.SetDescriptionf(b.String())
	embed.SetColor(colour)
	embed.SetTimestamp(timestamp)
	embed.SetThumbnail(thumbnail)
	builder := discord.NewMessageCreateBuilder()
	builder.SetEmbeds(embed.Build())
	return builder
}
