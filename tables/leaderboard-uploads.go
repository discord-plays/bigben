package tables

import "time"

type LeaderboardUploads struct {
	Year    int       `xorm:"year pk unique"`
	Sent    *bool     `xorm:"sent"`
	Created time.Time `xorm:"created"`
	Updated time.Time `xorm:"updated"`
}
