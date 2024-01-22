FROM golang:alpine

WORKDIR /app

COPY . .

RUN mkdir dist

RUN go build -ldflags="-s -w" -o dist/logmyip-server main.go pages.go data.go

EXPOSE 52899

CMD [ "dist/logmyip-server" ]

STOPSIGNAL SIGTERM