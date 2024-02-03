# Panopticon

Generic imageboard archiver

## Requirements:
- Postgres
- Redis
- Golang (building from source)

## Features:
- Asynchronous task queue
- Redis cache to avoid unnecessary writes
- JSON API and web interface
- Multiple imageboard support
- SHA256 file hashes
- Scalable (designate instances of the archiver to only download media, threads from a specific board or fetch job tasks)

## Setup:
`docker compose docker-compose.yml`

Building from source:
```
cd api
go build
./api
cd archiver
go build
./archiver
```
Running without building:
```
go run ./api
go run ./archiver
```