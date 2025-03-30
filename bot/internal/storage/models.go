package storage

import "time"

type Voting struct {
	ID          int
	Name        string
	ChannelID   string
	Options     []*Option
	TotalVoices int
	UniqueUsers map[string]int
	Owner       string
	OwnerID     string
	ExpDate     *time.Time
	CreatedAt   *time.Time
	Closed      bool
}

type Option struct {
	ID     int
	Name   string
	Voices int
}
