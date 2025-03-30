FROM golang:1.23 AS dependencies
WORKDIR /bot
COPY /bot/go.mod .
COPY /bot/go.sum .
RUN go mod download
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    libssl-dev \
    pkg-config && \
    rm -rf /var/lib/apt/lists/*    


FROM dependencies AS build
WORKDIR /bot
COPY bot/. .
RUN go build -o bin/bot cmd/main.go


FROM debian:stable-slim
WORKDIR /bot
COPY --from=build /usr/lib/x86_64-linux-gnu/libssl.so.3 /usr/lib/x86_64-linux-gnu/
COPY --from=build /usr/lib/x86_64-linux-gnu/libcrypto.so.3 /usr/lib/x86_64-linux-gnu/
COPY --from=build /bot/bin/bot .
COPY /bot/scripts/wait-for-it.sh .
RUN chmod +x /bot/bot
CMD ["/bot/wait-for-it.sh","dbTarantool:3301","--","/bot/bot"]