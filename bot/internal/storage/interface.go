package storage

import "time"

type Storage interface {
	Create(channelID, name, userID, username string, date *time.Time, options []string) (int, error)
	AddVoice(channelID, userID string, votingID, optionNum int) error
	Get(channelID string, votingID int) (Voting, error) //todo
	GetAll(channelID string) ([]Voting, error)          //todo
	Close(channelID string, votingID int, userID string) error
	Delete(channelID string, votingID int, userID string) error
	GetErrBadVotingID() error
	GetErrBadChannelID() error
	GetErrBadOptionID() error
	GetErrClosedVoting() error
	GetErrNoAccess() error
}
