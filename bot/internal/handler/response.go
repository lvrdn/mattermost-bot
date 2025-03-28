package handler

import (
	"fmt"
	"mmbot/internal/storage"
)

func makeResponse(voting storage.Voting) string {

	var response string

	var status string
	if voting.Closed {
		status = "closed"
	} else {
		status = "running"
	}

	var expDate string
	if voting.ExpDate == nil {
		expDate = "unlimited"
	} else {
		expDate = voting.ExpDate.Format(timeLayout)
	}

	createdAt := voting.CreatedAt.Format(timeLayout)

	response = fmt.Sprintf("voting id [%d], owner [%s], status [%s]\ntotal voted [%d], created at [%s], exp date [%s]\n%s:\n",
		voting.ID,
		voting.Owner,
		status,
		voting.TotalVoices,
		createdAt,
		expDate,
		voting.Name)

	i := 1
	for _, option := range voting.Options {
		var result float32
		if voting.TotalVoices == 0 {
			result = 0
		} else {
			result = float32(option.Voices) / float32(voting.TotalVoices) * 100
		}

		response += fmt.Sprintf("%d.   %.2f%% (voted [%d])   %s\n", i, result, option.Voices, option.Name)
		i++
	}

	return response
}
