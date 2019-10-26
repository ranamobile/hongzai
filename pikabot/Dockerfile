FROM golang:1.13.2-stretch

RUN go get github.com/nlopes/slack github.com/joho/godotenv

COPY src /app

VOLUME [ "/app" ]
ENTRYPOINT [ "go", "run", "/app/pikabot.go" ]
EXPOSE 8080
