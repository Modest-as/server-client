var grpcClient = require('./grpc_client');

function main() {
    grpcClient.subscribe(1337, function(call, lastMessage) {
        call.write({ message: "START" })

        call.on('data', function (result) {
            console.log(result.data.number)
        })
    
        setTimeout(function () {
            call.write({ message: "END" });
        }, 5000000);
    });
}

main()