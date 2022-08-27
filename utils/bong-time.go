package utils

import (
	"github.com/MrMelon54/BigBen/tables"
	"sync"
	"time"
	"xorm.io/xorm"
)

type CurrentBong struct {
	Engine    *xorm.Engine
	Text      string
	StartTime time.Time
	EndTime   time.Time
	mapLock   *sync.RWMutex
	guilds    map[string]*GuildCurrentBong
	mChan     chan ClickInfo
	mDone     chan struct{}
}

type GuildCurrentBong struct {
	Lock        *sync.RWMutex
	Dirty       bool
	Emoji       string
	MessageId   string
	ClickIds    []string
	ClickNames  []string
	MessageTime time.Time
}

type ClickInfo struct {
	GuildId   string
	MessageId string
	UserId    string
	Name      string
	Time      time.Time
}

func NewCurrentBong(engine *xorm.Engine, text string, sTime, eTime time.Time) *CurrentBong {
	c := &CurrentBong{
		Engine:    engine,
		Text:      text,
		StartTime: sTime,
		EndTime:   eTime,
		mapLock:   &sync.RWMutex{},
		guilds:    make(map[string]*GuildCurrentBong),
		mChan:     make(chan ClickInfo, 256),
		mDone:     make(chan struct{}, 1),
	}
	go c.internalLoop()
	return c
}

func (c *CurrentBong) internalLoop() {
	for {
		select {
		case <-c.mDone:
			break
		case i := <-c.mChan:
			g := c.GuildMapItem(i.GuildId)
			g.Lock.Lock()
			used := false
			mt := g.MessageTime
			if g.MessageId == i.MessageId {
				for _, j := range g.ClickIds {
					if j == i.UserId {
						goto exitClickCheck
					}
				}
				g.ClickIds = append(g.ClickIds, i.UserId)
				g.ClickNames = append(g.ClickNames, i.Name)
				g.Dirty = true
				used = true
			}
		exitClickCheck:
			g.Lock.Unlock()
			if used {
				_, _ = c.Engine.Insert(&tables.BongLog{
					GuildId:          i.GuildId,
					UserId:           i.UserId,
					Timestamp:        i.Time,
					MessageTimestamp: mt,
				})
			}
		}
	}
}

func (c *CurrentBong) Kill() {
	close(c.mDone)
}

func (c *CurrentBong) GuildMapItem(guildId string) *GuildCurrentBong {
	c.mapLock.RLock()
	g := c.guilds[guildId]
	c.mapLock.RUnlock()
	return g
}

func (c *CurrentBong) RandomGuildData(all []tables.GuildSettings) {
	c.mapLock.Lock()
	for _, i := range all {
		c.guilds[i.GuildId] = &GuildCurrentBong{
			Lock:       &sync.RWMutex{},
			Emoji:      RandomEmoji(i.BongEmoji),
			ClickIds:   []string{},
			ClickNames: []string{},
		}
	}
	c.mapLock.Unlock()
}

func (c *CurrentBong) TriggerClick(guildId, messageId, userId, name string) {
	if c.mChan == nil {
		return
	}
	c.mChan <- ClickInfo{
		GuildId:   guildId,
		MessageId: messageId,
		UserId:    userId,
		Name:      name,
		Time:      time.Now(),
	}
}
