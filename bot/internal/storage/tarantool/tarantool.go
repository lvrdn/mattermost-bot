package tarantool

import (
	"context"
	"encoding/json"
	"fmt"
	"mmbot/internal/config"
	"mmbot/internal/storage"
	"mmbot/pkg/logger"
	"sync"
	"time"

	"github.com/tarantool/go-tarantool"
)

var (
	errBadVotingID  error = fmt.Errorf("no voting with this id")
	errNoVotings    error = fmt.Errorf("no votings in this channel")
	errBadOptionID  error = fmt.Errorf("no option with this id")
	errClosedVoting error = fmt.Errorf("voting with this id closed")
	errNoAccess     error = fmt.Errorf("no access to close this voting")
)

type tarantoolStorage struct {
	connection *tarantool.Connection
}

func NewStorage(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config) (*tarantoolStorage, error) {

	const methodPointer string = "tarantool.NewStorage"

	opts := tarantool.Opts{User: "storage", Pass: "pass"}
	conn, err := tarantool.Connect("dbTarantool:3301", opts)

	if err != nil {
		logger.Error("connection to tarantool db failed", methodPointer, "text error", err.Error())
		return nil, err
	}

	go func() {
		defer wg.Done()
		defer logger.Info("closing tarantool connection", methodPointer)
		<-ctx.Done()
		conn.Close()
	}()

	ts := &tarantoolStorage{
		connection: conn,
	}
	return ts, nil
}

func (ts *tarantoolStorage) Create(channelID, name, userID, username string, date *time.Time, options []string) (int, error) {

	votingsOptions := make([]*storage.Option, len(options))

	for i := range votingsOptions {
		votingsOptions[i] = &storage.Option{
			ID:     i + 1,
			Name:   options[i],
			Voices: 0,
		}
	}

	now := time.Now()

	newVoting := &storage.Voting{
		Name:        name,
		ChannelID:   channelID,
		Options:     votingsOptions,
		TotalVoices: 0,
		UniqueUsers: make(map[string]int),
		Owner:       username,
		OwnerID:     userID,
		ExpDate:     date,
		CreatedAt:   &now,
		Closed:      false,
	}

	data, err := json.Marshal(newVoting)

	if err != nil {
		return 0, err
	}

	resp, err := ts.connection.Insert("votings", []interface{}{nil, channelID, data})
	if err != nil {
		return 0, err
	}

	insertedData := resp.Data[0].([]interface{})
	insertedID := int(insertedData[0].(uint64))

	return insertedID, nil
}

func (ts *tarantoolStorage) AddVoice(channelID, userID string, votingID, optionID int) error {

	resp, err := ts.connection.Select("votings", "idx_vt_id", 0, 1, tarantool.IterEq, []interface{}{votingID})
	if err != nil {
		return err
	}

	if len(resp.Data) == 0 {
		return errBadVotingID
	}

	record := resp.Data[0].([]interface{})
	jsonData := record[2].([]byte)

	voting := storage.Voting{}

	err = json.Unmarshal(jsonData, &voting)
	if err != nil {
		return err
	}

	checkClosedByDate(&voting)

	if voting.Closed {
		return errClosedVoting
	}

	if voting.ChannelID != channelID {
		return errBadVotingID
	}

	optionIDfromStorage, exist := voting.UniqueUsers[userID]
	if !exist {
		voting.TotalVoices++
		voting.UniqueUsers[userID] = optionID
		for _, option := range voting.Options {
			if option.ID == optionID {
				option.Voices++
				break
			}
		}

	} else {
		if optionIDfromStorage != optionID {
			voting.UniqueUsers[userID] = optionID

			for _, option := range voting.Options {
				if option.ID == optionID {
					option.Voices++
				} else if option.ID == optionIDfromStorage {
					option.Voices--
				}
			}

		}
	}

	data, err := json.Marshal(voting)
	if err != nil {
		return err
	}

	_, err = ts.connection.Replace("votings", []interface{}{votingID, channelID, data})
	if err != nil {
		return err
	}

	return nil
}
func (ts *tarantoolStorage) Get(channelID string, votingID int) (storage.Voting, error) { //todo

	resp, err := ts.connection.Select("votings", "idx_vt_id", 0, 1, tarantool.IterEq, []interface{}{votingID})
	if err != nil {
		return storage.Voting{}, err
	}

	if len(resp.Data) == 0 {
		return storage.Voting{}, errBadVotingID
	}

	record := resp.Data[0].([]interface{})
	jsonData := record[2].([]byte)

	voting := storage.Voting{}

	err = json.Unmarshal(jsonData, &voting)
	if err != nil {
		return storage.Voting{}, err
	}

	if voting.ChannelID != channelID {
		return storage.Voting{}, errBadVotingID
	}

	voting.ID = votingID
	checkClosedByDate(&voting)
	return voting, nil

}
func (ts *tarantoolStorage) GetAll(channelID string) ([]storage.Voting, error) { //todo

	resp, err := ts.connection.Select("votings", "idx_chan_id", 0, 10000, tarantool.IterEq, []interface{}{channelID})
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, errNoVotings
	}

	votings := make([]storage.Voting, len(resp.Data))

	for i := range votings {
		record := resp.Data[i].([]interface{})
		jsonData := record[2].([]byte)

		voting := storage.Voting{}

		err = json.Unmarshal(jsonData, &voting)
		if err != nil {
			return nil, err
		}
		votingID := int(record[0].(uint64))

		voting.ID = votingID
		checkClosedByDate(&voting)
		votings[i] = voting
	}

	return votings, nil

}

func (ts *tarantoolStorage) Close(channelID string, votingID int, userID string) error {

	resp, err := ts.connection.Select("votings", "idx_vt_id", 0, 1, tarantool.IterEq, []interface{}{votingID})
	if err != nil {
		return err
	}

	if len(resp.Data) == 0 {
		return errBadVotingID
	}

	record := resp.Data[0].([]interface{})
	jsonData := record[2].([]byte)

	voting := storage.Voting{}

	err = json.Unmarshal(jsonData, &voting)
	if err != nil {
		return err
	}

	checkClosedByDate(&voting)

	if voting.Closed {
		return errClosedVoting
	}

	if voting.OwnerID != userID {
		return errNoAccess
	}

	if voting.ChannelID != channelID {
		return errBadVotingID
	}

	voting.Closed = true

	data, err := json.Marshal(voting)
	if err != nil {
		return err
	}

	_, err = ts.connection.Replace("votings", []interface{}{votingID, channelID, data})
	if err != nil {
		return err
	}

	return nil

}
func (ts *tarantoolStorage) Delete(channelID string, votingID int, userID string) error {

	resp, err := ts.connection.Select("votings", "idx_vt_id", 0, 1, tarantool.IterEq, []interface{}{votingID})
	if err != nil {
		return err
	}

	if len(resp.Data) == 0 {
		return errBadVotingID
	}

	record := resp.Data[0].([]interface{})
	jsonData := record[2].([]byte)

	voting := storage.Voting{}

	err = json.Unmarshal(jsonData, &voting)
	if err != nil {
		return err
	}

	if voting.OwnerID != userID {
		return errNoAccess
	}

	if voting.ChannelID != channelID {
		return errBadVotingID
	}

	_, err = ts.connection.Delete("votings", "idx_vt_id", []interface{}{votingID})
	if err != nil {
		return err
	}

	return nil

}

func (ts *tarantoolStorage) GetErrBadVotingID() error {
	return errBadVotingID
}

func (ts *tarantoolStorage) GetErrNoVotings() error {
	return errNoVotings
}

func (ts *tarantoolStorage) GetErrBadOptionID() error {
	return errBadOptionID
}

func (ts *tarantoolStorage) GetErrClosedVoting() error {
	return errClosedVoting
}

func (ts *tarantoolStorage) GetErrNoAccess() error {
	return errClosedVoting
}

func checkClosedByDate(voting *storage.Voting) {
	if voting.ExpDate != nil {
		if time.Now().After(*voting.ExpDate) {
			voting.Closed = true
		}
	}
}
