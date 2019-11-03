FROM golang:1.13.2-stretch

COPY . /app
ENV GOPATH /app
WORKDIR /app

RUN go get github.com/nlopes/slack
RUN go get -d ./...

VOLUME [ "/app" ]
ENTRYPOINT [ "go", "run", "/app/src/pikabot.go" ]
EXPOSE 8080
