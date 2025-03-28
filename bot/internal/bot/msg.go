package bot

import (
	"fmt"
	"mmbot/pkg/logger"

	"github.com/mattermost/mattermost-server/v6/model"
)

const (
	initialMsg string = "hello! this is vote-bot\nuse \"help\" command to see commands list"
	byeMsg     string = "goodbye! I will be some later"
)

func sendMsg(msg string, channelID string, client *model.Client4) {

	const methodPointer string = "bot.SendMsg"

	post := &model.Post{}
	post.ChannelId = channelID
	post.Message = msg

	if _, _, err := client.CreatePost(post); err != nil {
		logger.Error("send msg error", methodPointer, "text error", err.Error())
	} else {
		msgForLog := fmt.Sprintf("msg [%s] sended", msg)
		logger.Info(msgForLog, methodPointer)
	}
}
