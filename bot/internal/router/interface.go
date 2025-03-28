package router

type CommandRouter interface {
	Route(command, channelID, userID, username string) string
}
