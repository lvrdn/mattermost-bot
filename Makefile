mattermost:
	docker-compose up -d

bot_db:
	docker-compose --profile disabled up -d


.PHONY: bot
bot:
	set -a; \
	. ./dev.env; \
	cd ./bot; \
	go build -o ./bin/bot.exe ./cmd/main.go; \
	./bin/bot.exe


