class StatelessClient {
    constructor(n) {
        this._n = n;
        this._timeout = 10000;
        this._count = 0
        this._sum = 0
        this._testPass = true

        console.log(`N: ${this._n}`)
    }

    get timeout() {
        return this._timeout
    }

    callback(call, lastMessage) {
        // start if we haven't received anything
        if (lastMessage === undefined) {
            call.write({ message: `START` })
        // continue if we had data before
        } else if (lastMessage.data !== undefined) {
            call.write({ message: `CONTINUE ${lastMessage.data.number}` });
        }

        call.on('data', (result) => this.handleResponse(call, result));
    }

    handleResponse(call, result) {
        if (result.error !== undefined) {
            console.log(`Error: ${result.error.message}`)
            return
        }

        if (this._count == this._n) {
            return
        }

        var val = Number(result.data.number);
        
        if (this._prev !== undefined) {
            if (val != this._prev * 3) {
                this._testPass = false
            }
        }

        this._prev = val;

        this._sum += this._prev 
        this._count += 1

        console.log(`Response: ${result.data.number} | Count: ${this._count}`)

        if (this._count == this._n) {
            call.write({ message: "END" });
            console.log(`Sum: ${this._sum}`)
            if (this._n < 13) console.log(`Test passed: ${this._testPass}`)
        }
    }
}

module.exports = {
    StatelessClient
}