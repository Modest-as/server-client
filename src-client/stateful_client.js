var max_n = 0xffff

class StatefulClient {
    constructor() {
        this._n = this.getN();
        this._getId = this.getId();
        this._count = 0
        this._timeout = 10000;

        console.log(`N: ${this._n}`)
        console.log(`UUID: ${this._getId}`)
    }

    get timeout() {
        return this._timeout
    }

    callback(call, lastMessage) {
        // start if we haven't received anything
        if (lastMessage === undefined) {
            call.write({ message: `START ${this._getId} ${this._n}` })
            // continue if we had data before
        } else if (lastMessage.data !== undefined) {
            call.write({ message: `CONTINUE ${this._getId}` });
        }

        call.on('data', (result) => this.handleResponse(call, result));
    }

    handleResponse(call, result) {
        if (result.error !== undefined) {
            console.log(`Error: ${result.error.message}`)
            return
        }

        this._sum += Number(result.data.number)
        this._count += 1

        if (this._count == this._n + 1) {
            console.log(`Checksum: ${result.data.number}`)
        } else {
            console.log(`Response: ${result.data.number} | Count: ${this._count}`)
        }
    }

    getN() {
        return Math.round(Math.random() * (max_n - 1) + 1);
    }

    getId() {
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
            var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
            return v.toString(16);
        });
    }
}

module.exports = {
    StatefulClient
}