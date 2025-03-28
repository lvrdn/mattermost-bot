package router

import (
	"mmbot/internal/handler"
	"mmbot/internal/storage"
	"strconv"
	"strings"
)

type router struct {
	botName string
	storage storage.Storage
}

func NewRouter(botName string, storage storage.Storage) *router {
	router := &router{
		botName: botName,
		storage: storage,
	}
	return router
}

func (r *router) Route(command, channelID, userID, username string) string {

	handler := handler.NewHandler(r.storage)

	lines := strings.Split(command, "\n")

	params := strings.Split(lines[0], " ")

	// if len(params) == 1 {
	// 	response := handler.NoCommand(r.botName)
	// 	return response
	// }

	switch params[1] {
	case "new": //todo lines[1,2,3...]

		if len(lines) < 3 || len(params) > 3 {
			return "bad input for \"new\" command, input \"help\" command for showing input params"
		}

		var expDate string
		if len(params) == 3 {
			expDate = params[2]
		}

		votingName := lines[1]

		options := lines[2:]

		response := handler.New(channelID, votingName, userID, username, expDate, options)

		return response

	case "vote":
		if len(params) != 4 || len(lines) != 1 {
			return "bad input for \"vote\" command, input \"help\" command for showing input params"
		}

		votingID, err := strconv.Atoi(params[2])
		if err != nil {
			return "voting id must be number"
		}
		optionNum, err := strconv.Atoi(params[3])
		if err != nil {
			return "voting option id must be number"
		}

		response := handler.Vote(channelID, userID, votingID, optionNum)

		return response

	case "show":
		if len(params) != 3 || len(lines) != 1 {
			return "bad input for \"show\" command, input \"help\" command for showing input params"
		}

		votingID, err := strconv.Atoi(params[2])
		if err != nil {
			return "voting id must be number"
		}

		response := handler.Show(channelID, votingID)

		return response

	case "show_all":
		if len(params) != 2 || len(lines) != 1 {
			return "bad input for \"show_all\" command, input \"help\" command for showing input params"
		}

		response := handler.ShowAll(channelID)

		return response

	case "close":
		if len(params) != 3 || len(lines) != 1 {
			return "bad input for \"close\" command, input \"help\" command for showing input params"
		}

		votingID, err := strconv.Atoi(params[2])
		if err != nil {
			return "voting id must be number"
		}

		response := handler.Close(channelID, votingID, userID)

		return response

	case "delete":
		if len(params) != 3 || len(lines) != 1 {
			return "bad input for \"delete\" command, input \"help\" command for showing input params"
		}

		votingID, err := strconv.Atoi(params[2])
		if err != nil {
			return "voting id must be number"
		}

		response := handler.Delete(channelID, votingID, userID)

		return response

	case "help":
		response := handler.Help(r.botName)
		return response
	default:
		return "unknown command, use \"help\" command to see commands list"
	}

}
