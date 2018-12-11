# Server Client Communication Example

## Dev notes

* Regenerate protobuf: `protoc -I grpc/ grpc/comms.proto --go_out=plugins=grpc:grpc`

## Run server

* Run server: `go run src-server/server.go -port [port] -mode [mode]`

**Available modes:** `stateless`, `stateful`

## Run client 

* Run client: `node main.js -p [port] -m [mode] [args]`

**Available modes:** `stateless`, `stateful`

**Args:** In stateless mode there should be one argument `n`

# Protocol

## Stateless

<Placeholder>

## Stateful

<Placeholder>

# Assumptions

<Placeholder>
