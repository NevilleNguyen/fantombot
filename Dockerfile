FROM golang:1.16-stretch AS build-env

ENV GO111MODULE=on
COPY . /go/src/github.com/quangkeu95/fantom-bot
WORKDIR /go/src/github.com/quangkeu95/fantom-bot
RUN go build

# FROM debian:stretch
# COPY --from=build-env /go/src/github.com/quangkeu95/fantom-bot/fantom-bot /

# RUN apt-get update && \
#     apt-get install -y ca-certificates && \
#     rm -rf /var/lib/apt/lists/*

CMD ["/go/src/github.com/quangkeu95/fantom-bot"]
