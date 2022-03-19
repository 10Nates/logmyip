# logmyip

A voluntary IP logging website. 
You can visit at https://logmyip.com.

## Compiling

Simple compile command to dist/ folder
`go build go build -ldflags="-s -w" -o dist/logmyip-server main.go pages.go data.go`

## Running

The environment variable "RedisPass" is required to run.
`RedisPass=MyPassword ./logmyip-server`