FROM golang:alpine

WORKDIR /app

COPY . .

RUN mkdir dist

RUN go build -ldflags="-s -w" -o dist/logmyip-server main.go pages.go data.go

# source no longer necessary
RUN rm main.go | rm pages.go | rm data.go | rm go.mod | rm go.sum 

EXPOSE 52899

CMD [ "dist/logmyip-server" ]

STOPSIGNAL SIGTERM