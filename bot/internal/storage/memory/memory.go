package memory

import (
	"context"
	"fmt"
	"mmbot/internal/config"
	"mmbot/internal/storage"
	"mmbot/pkg/logger"
	"sync"
	"time"
)

var (
	errBadVotingID  error = fmt.Errorf("no voting with this id")
	errNoVotings    error = fmt.Errorf("no votings in this channel")
	errBadOptionID  error = fmt.Errorf("no option with this id")
	errClosedVoting error = fmt.Errorf("voting with this id closed")
	errNoAccess     error = fmt.Errorf("no access to close this voting")
)

type memoryStorage struct {
	mu           *sync.RWMutex
	votings      map[string]map[int]*storage.Voting
	nextVotingID int
}

func NewStorage(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config) (*memoryStorage, error) {

	const methodPointer string = "memory.NewStorage"

	ms := &memoryStorage{
		mu:           &sync.RWMutex{},
		votings:      make(map[string]map[int]*storage.Voting),
		nextVotingID: 1,
	}

	go func() {
		defer wg.Done()
		defer logger.Info("closing memory storage", methodPointer)
		<-ctx.Done()
	}()

	return ms, nil
}

func (ms *memoryStorage) Create(channelID, name, userID, username string, date *time.Time, options []string) (int, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	votingOptions := make([]*storage.Option, len(options))
	for i := range votingOptions {
		option := &storage.Option{
			ID:   i + 1,
			Name: options[i],
		}
		votingOptions[i] = option
	}

	now := time.Now()

	newVoting := &storage.Voting{
		ID:          ms.nextVotingID,
		Name:        name,
		ChannelID:   channelID,
		Options:     votingOptions,
		UniqueUsers: make(map[string]int),
		Owner:       username,
		OwnerID:     userID,
		ExpDate:     date,
		CreatedAt:   &now,
	}

	ms.nextVotingID++

	if ms.votings[channelID] == nil {
		ms.votings[channelID] = make(map[int]*storage.Voting)
	}

	ms.votings[channelID][newVoting.ID] = newVoting

	return newVoting.ID, nil
}
func (ms *memoryStorage) AddVoice(channelID, userID string, votingID, optionID int) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	votingsStorage, exist := ms.votings[channelID]
	if !exist {
		ms.votings[channelID] = make(map[int]*storage.Voting)
	}

	voting, exist := votingsStorage[votingID]
	if !exist {
		return errBadVotingID
	}

	if optionID < 1 || optionID > len(voting.Options) {
		return errBadOptionID
	}

	checkClosedByDate(voting)

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

	return nil
}
func (ms *memoryStorage) Get(channelID string, votingID int) (storage.Voting, error) { //todo
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	votingsStorage, exist := ms.votings[channelID]
	if !exist {
		ms.votings[channelID] = make(map[int]*storage.Voting)
	}

	voting, exist := votingsStorage[votingID]
	if !exist {
		return storage.Voting{}, errBadVotingID
	}

	gettedVoting := *voting

	checkClosedByDate(&gettedVoting)

	return gettedVoting, nil
}
func (ms *memoryStorage) GetAll(channelID string) ([]storage.Voting, error) { //todo
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	votingsStorage, exist := ms.votings[channelID]
	if !exist {
		ms.votings[channelID] = make(map[int]*storage.Voting)
	}

	if len(votingsStorage) == 0 {
		return nil, errNoVotings
	}

	votings := make([]storage.Voting, len(votingsStorage))

	i := 0
	for _, voting := range votingsStorage {
		votings[i] = *voting

		checkClosedByDate(&votings[i])

		i++
	}
	return votings, nil

}

func (ms *memoryStorage) Close(channelID string, votingID int, userID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	votingsStorage, exist := ms.votings[channelID]
	if !exist {
		ms.votings[channelID] = make(map[int]*storage.Voting)
	}

	voting, exist := votingsStorage[votingID]
	if !exist {
		return errBadVotingID
	}

	if voting.OwnerID != userID {
		return errNoAccess
	}

	checkClosedByDate(voting)

	if voting.Closed {
		return errClosedVoting
	}

	if voting.ChannelID != channelID {
		return errBadVotingID
	}

	voting.Closed = true

	return nil
}
func (ms *memoryStorage) Delete(channelID string, votingID int, userID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	votingsStorage, exist := ms.votings[channelID]
	if !exist {
		ms.votings[channelID] = make(map[int]*storage.Voting)
	}

	voting, exist := votingsStorage[votingID]
	if !exist {
		return errBadVotingID
	}

	if voting.OwnerID != userID {
		return errNoAccess
	}
	if voting.ChannelID != channelID {
		return errBadVotingID
	}

	delete(votingsStorage, votingID)

	return nil
}

func (ms *memoryStorage) GetErrBadVotingID() error {
	return errBadVotingID
}

func (ms *memoryStorage) GetErrNoVotings() error {
	return errNoVotings
}

func (ms *memoryStorage) GetErrBadOptionID() error {
	return errBadOptionID
}

func (ms *memoryStorage) GetErrClosedVoting() error {
	return errClosedVoting
}

func (ms *memoryStorage) GetErrNoAccess() error {
	return errClosedVoting
}

func checkClosedByDate(voting *storage.Voting) {
	if voting.ExpDate != nil {
		if voting.ExpDate.Unix() < time.Now().Unix() {
			voting.Closed = true
		}
	}
}
