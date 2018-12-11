var grpcClient = require('./grpc_client');
var parseArgs = require('minimist');

var { StatelessClient } = require('./stateless_client');
var { StatefulClient } = require('./stateful_client');

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

    var client = undefined;

    if (mode == "stateless") {
        var n = Math.round(args["_"][0]);
        if (!(n > 0 && n <= 0xffff)) {
            console.log(
                `In stateless mode number of integers to ` +
                `receive must be specified between 1 and 0xffff`
            );
            process.exit();
        }
        client = new StatelessClient(n);
    } else if (mode == "stateful") {
        client = new StatefulClient();
    } else {
        console.log(`Mode should be either "statefull" or "stateless"`)
        process.exit();
    }

    grpcClient.subscribe(
        port = port,
        client.callback.bind(client),
        timeout = client.timeout
    );
}

main();