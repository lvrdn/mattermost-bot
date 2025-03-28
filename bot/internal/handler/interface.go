package handler

type Handler interface {
	NoCommand(botName string) string

	New(channelID, name, userID, username, expDate string, options []string) string

	Vote(channelID, userID string, votingID, optionNum int) string

	Show(channelID string, votingID int) string

	ShowAll(channelID string) string

	Close(channelID string, votingID int, userID string) string

	Delete(channelID string, votingID int, userID string) string

	Help(botName string) string
}
