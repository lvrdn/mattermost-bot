package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"mmbot/internal/config"
	"mmbot/internal/router"
	"mmbot/internal/storage/tarantool"
	"mmbot/pkg/logger"
	"os"
	"os/signal"
	"strings"
	"sync"

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
	wg                        *sync.WaitGroup
	channelsID                sync.Map
}

func NewBot(ctx context.Context, cfg *config.Config) *bot {

	const methodPointer string = "bot.NewBot"

	bot := &bot{
		config: cfg,
		wg:     &sync.WaitGroup{},
	}

	bot.wg.Add(1)
	storage, err := tarantool.NewStorage(ctx, bot.wg, cfg)
	if err != nil {
		logger.Fatal("connection to storage failed", methodPointer, "text error", err.Error())
	}

	bot.router = router.NewRouter(cfg.MMBotname, storage)

	// create a new mattermost client
	bot.mattermostClient = model.NewAPIv4Client(bot.config.MMServer.String())

	// login
	bot.mattermostClient.SetToken(cfg.MMToken)

	if user, resp, err := bot.mattermostClient.GetUser("me", ""); err != nil {
		logger.Fatal("get user failed", methodPointer, "text error", err.Error())
	} else {
		logger.Info("get user successfully", methodPointer, "resp", resp)
		bot.mattermostUser = user
	}

	// find and save the bot's team to app struct
	if team, resp, err := bot.mattermostClient.GetTeamByName(cfg.MMTeam, ""); err != nil {
		logger.Fatal("get team failed", methodPointer, "text error", err.Error())
	} else {
		logger.Info("get team successfully", methodPointer, "resp", resp)
		bot.mattermostTeam = team
	}

	// find and save the talking channel to app struct
	for _, channelName := range cfg.MMChannels {
		if channel, resp, err := bot.mattermostClient.GetChannelByName(
			channelName, bot.mattermostTeam.Id, "",
		); err != nil {
			logger.Warn("get channel failed", methodPointer, "channel name", channelName, "text error", err.Error())
		} else {
			logger.Info("get channel successfully", methodPointer, "channel name", channelName, "resp", resp)
			bot.mattermostChannels = append(bot.mattermostChannels, channel)
			bot.channelsID.Store(channel.Id, struct{}{})
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

func (b *bot) ListenToEvents(ctx context.Context) {

	const methodPointer string = "bot.ListenToEvents"

	var err error

	url := fmt.Sprintf("ws://%s", b.config.MMServer.Host+b.config.MMServer.Path)
	b.mattermostWebSocketClient, err = model.NewWebSocketClient4(
		url,
		b.config.MMToken,
	)
	if err != nil {
		logger.Fatal("mattermost websocket disconnected, stopped", methodPointer)
	}

	logger.Info("mattermost websocket connected", methodPointer)

	b.mattermostWebSocketClient.Listen()

	for event := range b.mattermostWebSocketClient.EventChannel {
		b.wg.Add(1)
		go b.handleWebSocketEvent(event)
	}

}

func (b *bot) GracefulShutdown() {

	const methodPointer string = "bot.GracefulShutdown"

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	for _, channel := range b.mattermostChannels {
		sendMsg(byeMsg, channel.Id, b.mattermostClient)
	}

	if b.mattermostWebSocketClient != nil {
		logger.Info("closing websocket connection", methodPointer)
		b.mattermostWebSocketClient.Close()
	}

}

func (b *bot) WaitСlosingProcesses() {

	const methodPointer string = "bot.Wait"

	defer logger.Info("Shutting down", methodPointer)
	b.wg.Wait()
}

func (b *bot) handleWebSocketEvent(event *model.WebSocketEvent) {

	defer b.wg.Done()

	const methodPointer string = "bot.handleWebSocketEvent"

	// Ignore other types of events.
	if event.EventType() != model.WebsocketEventPosted {
		return
	}

	// Ignore other channels
	if _, ok := b.channelsID.Load(event.GetBroadcast().ChannelId); !ok {
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
