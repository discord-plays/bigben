package bigben

import (
	"fmt"
	"github.com/MrMelon54/BigBen/commands"
	"github.com/MrMelon54/BigBen/tables"
	"github.com/bwmarrin/discordgo"
	"xorm.io/xorm"
)

type BigBen struct {
	engine          *xorm.Engine
	appId           string
	guildId         string
	session         *discordgo.Session
	commands        commands.CommandList
	commandHandlers map[string]commands.CommandHandler
}

func (b *BigBen) AppId() string { return b.appId }

func (b *BigBen) GuildId() string             { return b.guildId }
func (b *BigBen) Session() *discordgo.Session { return b.session }
func NewBigBen(engine *xorm.Engine, token, appId, guildId string) (*BigBen, error) {
	bot, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	bot.AddHandler(func(s *discordgo.Session, _ *discordgo.Ready) {
		_ = s.UpdateGameStatus(0, "Bong")
	})
	bot.Identify.Intents = discordgo.IntentGuilds | discordgo.IntentGuildMessages
	err = bot.Open()
	if err != nil {
		return nil, err
	}
	return (&BigBen{}).init(engine, appId, guildId, bot)
}

func (b *BigBen) init(engine *xorm.Engine, appId, guildId string, bot *discordgo.Session) (*BigBen, error) {
	b.engine = engine
	b.appId = appId
	b.guildId = guildId
	b.session = bot
	b.commands, b.commandHandlers = commands.InitCommands(b)
	err := b.updateCommands()
	if err != nil {
		return nil, err
	}
	bot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := b.commandHandlers[i.ApplicationCommandData().Name]; ok {
			h.Handler(s, i)
		}
	})
	return b, nil
}

func (b *BigBen) Exit() error {
	return b.session.Close()
}

func (b *BigBen) updateCommands() error {
	_, err := b.session.ApplicationCommandBulkOverwrite(b.appId, b.guildId, b.commands)
	if err != nil {
		return fmt.Errorf("bulk overwrite application commands error: %s", err)
	}
	return nil
}

func (b *BigBen) GetGuildSettings(guildId string) (tables.GuildSettings, error) {
	var g tables.GuildSettings
	_, err := b.engine.Where("guild_id = ?", guildId).Get(&g)
	return tables.GuildSettings{GuildId: guildId}, err
}

func (b *BigBen) PutGuildSettings(guildSettings tables.GuildSettings) error {
	ok, err := b.engine.Where("guild_id = ?", guildSettings.GuildId).Update(&guildSettings)
	if ok == 0 || err != nil {
		_, err = b.engine.Insert(&guildSettings)
		return err
	}
	return nil
}
