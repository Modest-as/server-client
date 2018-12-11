var PROTO_PATH = __dirname + '/../grpc/comms.proto';

var grpc = require('grpc');
var protoLoader = require('@grpc/proto-loader');

var packageDefinition = protoLoader.loadSync(
    PROTO_PATH,
    {
        keepCase: true,
        longs: String,
        enums: String,
        defaults: true,
        oneofs: true
    });

var server = grpc.loadPackageDefinition(packageDefinition).grpc;

var backoff_min_value = 500
var backoff_rate = 2
var backoff = 500

// exposes a way to inject behaviour to the bidirectional
// GRPC stream and abstracts exponential backoff implementation
function subscribe(port, action) {
    var client = new server.Comms(`localhost:${port}`, grpc.credentials.createInsecure());

    var interceptor = function (options, nextCall) {
        var lastMessage;

        var requester = {
            start: function(metadata, _, next) {
                var newListener = {
                    onReceiveMessage: function(message, next) {
                        lastMessage = message
                        backoff = backoff_min_value
                        next(message);
                    },
                    onReceiveStatus: function(status, next) {
                        if (status.code !== grpc.status.OK) {
                            backoff *= backoff_rate
                            console.log(`Failed to connect, will retry in ${backoff} seconds`)
                            setTimeout(function () {
                                var call = client.GetNumbers({ interceptors: [interceptor] });
                                action(call, lastMessage)
                            }, backoff);
                            return
                        }
                        
                        next(status);
                    }
                };
    
                next(metadata, newListener);
            }
        };
    
        return new grpc.InterceptingCall(nextCall(options), requester);
    };

    var call = client.GetNumbers({ interceptors: [interceptor] });

    action(call, null)
}

module.exports = {
    subscribe
}