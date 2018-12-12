# Server Client Communication Example

## Dev notes

* Regenerate protobuf: `protoc -I grpc/ grpc/comms.proto --go_out=plugins=grpc:grpc`

## Run server

* Run server: `go run src-server/server.go -port [port] -mode [mode]`

**Available modes:** `stateless`, `stateful`

## Run client 

* Run client: `node src-client/main.js -p [port] -m [mode] [args]`

**Available modes:** `stateless`, `stateful`, `stateful-test`, `stateful-test-reconnect`

`stateful-test` and `stateful-test-reconnect` are used for testing, they have fixed `uuid` and `n` values. Stateful reconnects can be tested by terminating client when in `stateful-test` mode and then starting it again in `stateful-test-reconnect` mode.

**Args:** In stateless mode there should be one argument `n`

# Features

* Client is using exponential backoff to prevent busy-looping when connection drops.
* Stateful server discards abandoned sessions after 30 seconds of inactivity
* Server accepts port values
* Server can handle multiple clients
* Server can switch between stateless/stateful modes
* Client accepts custom server port
* Client can switch between stateless/stateful modes
* Helpful error messages from the server in order to debug protocol issues

# Protocol

## Stateless

Client connects the the server and sends a GRPC comms message `START` with no parameters to initialize stream.

Once server receives `START` message it will generate a random number `a` between `1` and `0xff` and will start sending them one by one as GRPC `Reply.Data` messages each time multiplying `a` by `3`. Integer overflows are not handled and the number is allowed to wrap around. 

Server will continue sending messages until client sends `END` message to the server. It is clients responsibility to determine how many messages it needs.

In case connection drops or server dies, client can reconnect sending `CONTINUE [x]` message where `[x]` is 
the last message server delivered. Server then continues sending messages to the client from `x` onwards.

## Stateful

Client connects the the server and sends a GRPC comms message `START [uuid] [n]` where `[uuid]` is unique session identifier and `[n]` is a session specific value of how many messages server should send back.

Once server receives `START [uuid] [n]` it will store this session id information on the server side and will start sending random `uint32` numbers. Once all numbers are sent server will close the stream.

In case the connection drops, client can reconnect by sending `CONTINUE [uuid]` message. Note that the server saves session information for 30 seconds before discarding so if client doesn't reconnect after `30` seconds, the continue attempt will fail and server will respond with an error message.

When client reconnects server will send two integers before continuing with the normal stream:

1) Amount of numbers already delivered 
2) Checksum of numbers already delivered 

Random data that is sent is deterministic and and the seed is calculated based on first 8 bytes of the UUID that is sent to the server.

UUID to Seed mapping:

Take first 8 bytes of the UUID -> Assume that they are in big endian configuration -> convert that to 64 bit integer.

We are using `math/rand` package to generate random integers. 

The minimum required state that we need to keep track of to achieve this functionality is the UUID, the total desired amount of numbers and current amount. We map UUID to seed using the mapping described above and we can calculate n-th number that we want to send just based on that.

The checksum algorithm is very straightforward modular sum base `123456`.

# Assumptions

* GRPC ensures that the bidirectional streaming request are ordered
* Server doesn't guarantee that all client messages get executed, if the client sends two or more messages in quick succession (less than a second apart) server will only respond to the last received message.
* UUID is unique for every value `n`. I.e. there is no situation where two sessions want to be initialized simultaneously `(id, n1)` and `(id, n2)` and `n1 != n2`. 


# Testing 

No explicit testing framework implemented. There are `stateful-test` and `stateful-test-reconnect` modes in stateful configuration that allow testing connect and reconnect behaviors in the stateful mode and server can be manually disabled in the stateless mode to test the same behavior.

Happy path gets tested on the client side with n values less than 13 (didn't want to replicate wrapping behavior in Node) and checksum gets calculated and verified against the server version when in stateful 
configuration.