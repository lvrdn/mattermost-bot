package bot

import (
	"encoding/json"
	"fmt"
	"mmbot/internal/config"
	"mmbot/internal/router"
	"mmbot/internal/storage/memory"
	"mmbot/pkg/logger"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/v6/model"
)

type bot struct {
	router                    router.CommandRouter
	config                    *config.Config
	mattermostClient          *model.Client4
	mattermostWebSocketClient *model.WebSocketClient
	mattermostUser            *model.User
	mattermostChannels        []*model.Channel
	mattermostTeam            *model.Team
}

func NewBot(cfg *config.Config) *bot {

	const methodPointer string = "bot.NewBot"

	storage := memory.NewStorage(cfg)

	bot := &bot{
		config: cfg,
		router: router.NewRouter(cfg.MMBotname, storage),
	}

	// Create a new mattermost client.
	bot.mattermostClient = model.NewAPIv4Client(bot.config.MMServer.String())

	// Login.
	bot.mattermostClient.SetToken(cfg.MMToken)

	if user, resp, err := bot.mattermostClient.GetUser("me", ""); err != nil {
		logger.Fatal("get user failed", methodPointer, "text error", err.Error())
	} else {
		logger.Info("get user successfully", methodPointer, "resp", resp)
		bot.mattermostUser = user
	}

	// Find and save the bot's team to app struct.
	if team, resp, err := bot.mattermostClient.GetTeamByName(cfg.MMTeam, ""); err != nil {
		logger.Fatal("get team failed", methodPointer, "text error", err.Error())
	} else {
		logger.Info("get team successfully", methodPointer, "resp", resp)
		bot.mattermostTeam = team
	}

	// Find and save the talking channel to app struct.
	for _, channelName := range cfg.MMChannels {
		if channel, resp, err := bot.mattermostClient.GetChannelByName(
			channelName, bot.mattermostTeam.Id, "",
		); err != nil {
			logger.Warn("get channel failed", methodPointer, "channel name", channelName, "text error", err.Error())
		} else {
			logger.Info("get channel successfully", methodPointer, "channel name", channelName, "resp", resp)
			bot.mattermostChannels = append(bot.mattermostChannels, channel)
		}
	}

	if bot.mattermostChannels == nil {
		logger.Fatal("no channels received", methodPointer)
	}

	for _, channel := range bot.mattermostChannels {
		sendMsg(initialMsg, channel.Id, bot.mattermostClient)
	}

	return bot
}

func (b *bot) ListenToEvents() {

	const methodPointer string = "bot.ListenToEvents"

	var err error
	tryNum := 1
	for {
		url := fmt.Sprintf("ws://%s", b.config.MMServer.Host+b.config.MMServer.Path)
		b.mattermostWebSocketClient, err = model.NewWebSocketClient4(
			url,
			b.config.MMToken,
		)
		if err != nil {
			logger.Warn("mattermost websocket disconnected, retrying", methodPointer, "tryNum", tryNum, "text error", err.Error())
			tryNum += 1

			if tryNum >= 5 {
				logger.Fatal("mattermost websocket disconnected, stopped", methodPointer)
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}
		logger.Info("Mattermost websocket connected", methodPointer)

		b.mattermostWebSocketClient.Listen()

		for event := range b.mattermostWebSocketClient.EventChannel {
			go b.handleWebSocketEvent(event)
		}

	}

	// for _, channel := range b.MattermostChannels {
	// 	sendMsg(byeMsg, channel, b.MattermostClient)
	// }

}

func (b *bot) SetupGracefulShutdown() {

	const methodPointer string = "bot.SetupGracefulShutdown"

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			if b.mattermostWebSocketClient != nil {
				logger.Info("Closing websocket connection", methodPointer)
				b.mattermostWebSocketClient.Close()
			}
			logger.Info("Shutting down", methodPointer)
			os.Exit(0)
		}
	}()
}

func (b *bot) handleWebSocketEvent(event *model.WebSocketEvent) {

	const methodPointer string = "bot.handleWebSocketEvent"

	// Ignore other types of events.
	if event.EventType() != model.WebsocketEventPosted {
		return
	}

	// Ignore other channels.
	ok := false
	for _, channel := range b.mattermostChannels { //data race!
		if event.GetBroadcast().ChannelId == channel.Id {
			ok = true
			break
		}
	}
	if !ok {
		return
	}

	// unmarshal event to (*model.Post)
	post := &model.Post{}
	err := json.Unmarshal([]byte(event.GetData()["post"].(string)), &post)
	if err != nil {
		logger.Error("unmarshall post failed", methodPointer, "error text", err.Error())
		return
	}

	// Ignore messages sent by this bot itself.
	if post.UserId == b.mattermostUser.Id {
		return
	}

	// Ignore messages without calling @bot in the beggining of message
	if prefix := fmt.Sprintf("@%s", b.config.MMBotname); !strings.HasPrefix(post.Message, prefix) {
		return
	}
	user, _, err := b.mattermostClient.GetUser(post.UserId, "")
	if err != nil {
		logger.Error("get user error", methodPointer, "error text", err.Error())
		sendMsg("something went wrong, please try later", post.ChannelId, b.mattermostClient)
		return
	}
	response := b.router.Route(post.Message, post.ChannelId, post.UserId, user.Username)

	if response != "" {
		sendMsg(response, post.ChannelId, b.mattermostClient)
	}

}
