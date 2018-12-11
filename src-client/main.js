var PROTO_PATH = __dirname + '/../grpc/comms.proto';
var grpc = require('grpc');
var protoLoader = require('@grpc/proto-loader');

// Suggested options for similarity to existing grpc.load behavior
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

function main() {
    var client = new server.Comms('localhost:1337', grpc.credentials.createInsecure());

    var call = client.GetNumbers();

    call.write({ message: "START" })

    call.on('data', function (result) {
        console.log(result.data.number)
    })

    setTimeout(function () {
        call.write({ message: "END" });
    }, 5000000);
}

main()

