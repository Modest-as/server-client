var max_n = 0xfff
var checksum_mod = 123456

class StatefulClient {
    constructor(id = undefined, n = undefined, cnt = false) {
        this._n = n === undefined ? this.getN() : n;
        this._getId = id === undefined ? this.getId() : id;
        this._count = 0
        this._timeout = 10000;
        this._checksum = 0;
        this._cnt = cnt;

        console.log(`N: ${this._n}`)
        console.log(`UUID: ${this._getId}`)
    }

    get timeout() {
        return this._timeout
    }

    callback(call, lastMessage) {
        // start if we haven't received anything
        if (lastMessage === undefined && !this._cnt) {
            call.write({ message: `START "${this._getId}" ${this._n}` })
            // continue if we had data before
        } else if (this._cnt || lastMessage.data !== undefined) {
            call.write({ message: `CONTINUE ${this._getId}` });
            this._cnt = false
            this._storeCount = true
        }

        call.on('data', (result) => this.handleResponse(call, result));
    }

    handleResponse(call, result) {
        if (result.error !== undefined) {
            console.log(`Error: ${result.error.message}`)
            return;
        }

        if (this._storeCount) {
            this._storeCount = false;
            this._storeChecksum = true;
            this._count = Number(result.data.number);
            console.log(`Count set to: ${this._count}`)
            return;
        }

        if (this._storeChecksum) {
            this._storeChecksum = false;
            this._checksum = Number(result.data.number);
            console.log(`Checksum set to: ${this._checksum}`)
            return;
        }

        this._count += 1

        if (this._count == this._n + 1) {
            console.log(`Checksum: ${result.data.number}`)
            console.log(`Checksum is valid: ${result.data.number == this._checksum}`)
        } else {
            console.log(`Response: ${result.data.number} | Count: ${this._count}`)
        }

        this._checksum += Number(result.data.number) % checksum_mod
        this._checksum = this._checksum % checksum_mod
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