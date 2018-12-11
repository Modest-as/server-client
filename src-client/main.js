var grpcClient = require('./grpc_client');
var parseArgs = require('minimist');

var { StatelessClient } = require('./stateless_client');

function main() {
    var args = parseArgs(process.argv.slice(2));

    var mode = "stateless";

    if (args["m"] !== undefined) {
        mode = args["m"];
    }

    var port = 1337;

    if (args["p"] !== undefined) {
        port = args["p"];
    }

    console.log(`Mode: ${mode}`);
    console.log(`Port: ${port}`);

    if (mode == "stateless" && !(args["_"][0] > 0 && args["_"][0] <= 0xffff)) {
        console.log(
            `In stateless mode number of integers to ` +
            `receive must be specified between 1 and 0xffff`
        );
        process.exit();
    }

    var client = undefined;

    if (mode == "stateless") {
        client = new StatelessClient(args["_"][0]);
    }

    grpcClient.subscribe(
        port = 1337,
        client.callback.bind(client),
        timeout = client.timeout
    );
}

main();