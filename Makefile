mattermost:
	curl -O https://raw.githubusercontent.com/mattermost/docker/master/docker-compose.yml

run:
	docker-compose up -d

.PHONY: bot
bot:
	set -a; \
	. ./dev.env; \
	cd ./bot; \
	go build -o ./bin/bot.exe ./cmd/main.go; \
	./bin/bot.exe
