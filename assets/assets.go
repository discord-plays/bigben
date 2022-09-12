package assets

import (
	"embed"
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"io"
	"time"
)

//go:embed clock-faces
var clockFaces embed.FS

var ukLocation = func() *time.Location {
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		panic(err)
	}
	return location
}()

func ReadClockFace(i int) (io.Reader, error) {
	return clockFaces.Open(fmt.Sprintf("clock-faces/%d.png", i))
}

func ReadClockFaceAsOptionalIcon(i int) *discord.Icon {
	if face, err := ReadClockFace(i); err == nil {
		if icon, err := discord.NewIcon(discord.IconTypePNG, face); err == nil {
			return icon
		}
	}
	return nil
}

func ReadClockFaceByTimeAsOptionalIcon(t time.Time) *discord.Icon {
	return ReadClockFaceAsOptionalIcon(t.In(ukLocation).Hour() % 12)
}
