# logmyip

A voluntary IP logging website.
You can visit at <https://logmyip.com>.

## Docker

Everything is handled in Docker Compose
`REDIS_PASSWORD=MyPassword docker compose up`

This will also create a Redis instance and a volume to go with it. Data is stored persistently there.
Redis is not directly exposed to the user and you will have to enter the redis container to use redis-cli.

## Compiling (No Docker)

Simple compile command to dist/ folder
`go build -ldflags="-s -w" -o dist/logmyip-server main.go pages.go data.go`

## Running (No Docker)

A Redis instance is required as one is not started automatically in this way.

The environment variable "RedisPass" is required to run.
`RedisAddr=localhost:6379 RedisPass=MyPassword ./logmyip-server`
