package handler

import (
	"fmt"
	"mmbot/internal/storage"
	"mmbot/pkg/logger"
	"time"
)

const (
	timeLayout string = "15:04:05-02.01.2006"

	somethingWrong string = "something went wrong, please try later"

	noCommand string = "enter command after @%s, use \"help\" command to see commands list"

	noVotings string = "no votings in this channel now"

	helpResponse string = `
for use bot enter folowing commands with format:
- create voting - use minimum 3 lines: 
		1. 	@%s new optional_exp_date(hh:mm:ss-dd:mm:yyyy)
		2. 	voting name
		3. 	option1 name
		n. 	in next lines option2 name, option3 name, ...
- vote: 
		@%s vote vote_id(number) var(number)
- show voting with id: 
		@%s show vote_id(number)
- show all: 
		@%s show_all
- voting owner can close it: 
		@%s close vote_id(number)
- voting owner can delete it: 
		@%s close vote_id(number)
`
)

type handler struct {
	storage storage.Storage
}

func NewHandler(storage storage.Storage) Handler {
	return &handler{
		storage: storage,
	}
}

func (h *handler) NoCommand(botName string) string {

	response := fmt.Sprintf(noCommand, botName)

	return response
}

func (h *handler) New(channelID, name, userID, username, expDate string, options []string) string {

	const methodPointer string = "handler.New"

	var date *time.Time

	if expDate != "" {
		var err error
		date = new(time.Time)
		*date, err = time.ParseInLocation(timeLayout, expDate, time.Now().Location())
		if err != nil {
			return "enter time in next format [ss:mm:hh-dd:mm:yyyy]"
		}

	}

	votingID, err := h.storage.Create(channelID, name, userID, username, date, options)
	if err != nil {
		logger.Error("create new voting error", methodPointer, "text error", err.Error())
		return somethingWrong
	}

	response := fmt.Sprintf("voting created with id [%d]", votingID)

	return response
}

func (h *handler) Vote(channelID, userID string, votingID, optionID int) string {

	const methodPointer string = "handler.Vote"

	err := h.storage.AddVoice(channelID, userID, votingID, optionID)
	switch err {
	case h.storage.GetErrBadVotingID():
		response := fmt.Sprintf("no voting with this id [%d]", votingID)
		logger.Info("vote failed - bad voting id", methodPointer, "text error", err.Error(), "response", response)
		return response

	case h.storage.GetErrBadOptionID():
		response := fmt.Sprintf("no option with this id [%d]", optionID)
		logger.Info("vote failed - bad voting id", methodPointer, "text error", err.Error(), "response", response)
		return response

	case h.storage.GetErrClosedVoting():
		response := fmt.Sprintf("voting with this id [%d] closed", votingID)
		logger.Info("vote failed - voting closed", methodPointer, "text error", err.Error(), "response", response)
		return response

	default:
		if err != nil {
			logger.Error("vote failed", methodPointer, "text error", err.Error(), "channel id", channelID, "user id", userID, "voting id", votingID, "option id", optionID)
			return somethingWrong
		}

	}

	response := ""

	return response
}

func (h *handler) Show(channelID string, votingID int) string {

	const methodPointer string = "handler.Show"

	voting, err := h.storage.Get(channelID, votingID) //todo
	switch err {
	case h.storage.GetErrBadVotingID():
		response := fmt.Sprintf("no voting with this id [%d]", votingID)
		logger.Info("get voting failed - bad voting id", methodPointer, "text error", err.Error(), "response", response)
		return response

	default:
		if err != nil {
			logger.Error("get voting failed", methodPointer, "text error", err.Error(), "channel id", channelID, "voting id", votingID)
			return somethingWrong
		}

	}

	response := makeResponse(voting)

	return response
}

func (h *handler) ShowAll(channelID string) string {

	const methodPointer string = "handler.ShowAll"

	votings, err := h.storage.GetAll(channelID)
	switch err {
	case h.storage.GetErrNoVotings():
		response := "no votings in this channel now"
		logger.Info("get all votings failed - no votings with this channel id", methodPointer, "response", response)
		return response

	default:
		if err != nil {
			logger.Error("get all votings failed", methodPointer, "text error", err.Error(), "channel id", channelID)
			return somethingWrong
		}

	}

	var response string

	for _, voting := range votings {
		response += makeResponse(voting)
		response += "\n"
	}

	if response == "" {
		response = noVotings
	}

	return response
}

func (h *handler) Close(channelID string, votingID int, userID string) string {

	const methodPointer string = "handler.Close"

	err := h.storage.Close(channelID, votingID, userID) //todo
	switch err {
	case h.storage.GetErrBadVotingID():
		response := fmt.Sprintf("no voting with this id [%d]", votingID)
		logger.Info("close voting failed - bad voting id", methodPointer, "text error", err.Error(), "response", response)
		return response

	case h.storage.GetErrNoAccess():
		response := fmt.Sprintf("no access to close voting with id [%d]", votingID)
		logger.Info("close voting failed - no access", methodPointer, "text error", err.Error(), "response", response, "user id", userID, "voting id", votingID)
		return response

	case h.storage.GetErrClosedVoting():
		response := fmt.Sprintf("voting with this id [%d] closed", votingID)
		logger.Info("close voting failed - voting closed", methodPointer, "text error", err.Error(), "response", response)
		return response

	default:
		if err != nil {
			logger.Error("close voting failed", methodPointer, "text error", err.Error(), "channel id", channelID, "user id", userID, "voting id", votingID)
			return somethingWrong
		}

	}

	response := fmt.Sprintf("voting with id [%d] is closed", votingID)

	return response
}

func (h *handler) Delete(channelID string, votingID int, userID string) string {

	const methodPointer string = "handler.Delete"

	err := h.storage.Delete(channelID, votingID, userID)
	switch err {
	case h.storage.GetErrBadVotingID():
		response := fmt.Sprintf("no voting with this id [%d]", votingID)
		logger.Info("delete voting failed - bad voting id", methodPointer, "text error", err.Error(), "response", response)
		return response

	case h.storage.GetErrNoAccess():
		response := fmt.Sprintf("no access to delete voting with id [%d]", votingID)
		logger.Info("delete voting failed - no access", methodPointer, "text error", err.Error(), "response", response, "user id", userID, "voting id", votingID)
		return response

	default:
		if err != nil {
			logger.Error("delete voting failed", methodPointer, "text error", err.Error(), "channel id", channelID, "user id", userID, "voting id", votingID)
			return somethingWrong
		}

	}

	response := ""

	return response
}

func (h *handler) Help(botName string) string {

	response := fmt.Sprintf(helpResponse, botName, botName, botName, botName, botName, botName)

	return response
}
