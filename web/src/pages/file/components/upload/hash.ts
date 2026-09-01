const HASH_CHUNK_SIZE = 2 * 1024 * 1024;

class Md5 {
    private a = 0x67452301 | 0;
    private b = 0xefcdab89 | 0;
    private c = 0x98badcfe | 0;
    private d = 0x10325476 | 0;
    private buffer = new Uint8Array(64);
    private bufferLength = 0;
    private bytesHashed = 0;

    update(data: Uint8Array) {
        let position = 0;
        this.bytesHashed += data.length;

        if (this.bufferLength > 0) {
            const available = 64 - this.bufferLength;
            const used = Math.min(available, data.length);
            this.buffer.set(data.subarray(0, used), this.bufferLength);
            this.bufferLength += used;
            position += used;
            if (this.bufferLength === 64) {
                this.processBlock(this.buffer, 0);
                this.bufferLength = 0;
            }
        }

        while (position + 64 <= data.length) {
            this.processBlock(data, position);
            position += 64;
        }

        if (position < data.length) {
            this.buffer.set(data.subarray(position), 0);
            this.bufferLength = data.length - position;
        }
    }

    digest() {
        const bitLength = this.bytesHashed * 8;
        const paddingLength = this.bufferLength < 56 ? 56 - this.bufferLength : 120 - this.bufferLength;
        const padding = new Uint8Array(paddingLength + 8);
        const low = bitLength >>> 0;
        const high = Math.floor(bitLength / 0x100000000) >>> 0;

        padding[0] = 0x80;
        for (let i = 0; i < 4; i++) {
            padding[paddingLength + i] = (low >>> (8 * i)) & 0xff;
            padding[paddingLength + 4 + i] = (high >>> (8 * i)) & 0xff;
        }
        this.update(padding);

        return [this.a, this.b, this.c, this.d].map(wordToHex).join("");
    }

    private processBlock(data: Uint8Array, offset: number) {
        const x: number[] = [];
        for (let i = 0; i < 16; i++) {
            const index = offset + i * 4;
            x[i] = (data[index] | (data[index + 1] << 8) | (data[index + 2] << 16) |
                (data[index + 3] << 24)) | 0;
        }

        let a = this.a;
        let b = this.b;
        let c = this.c;
        let d = this.d;

        a = ff(a, b, c, d, x[0], 7, -680876936);
        d = ff(d, a, b, c, x[1], 12, -389564586);
        c = ff(c, d, a, b, x[2], 17, 606105819);
        b = ff(b, c, d, a, x[3], 22, -1044525330);
        a = ff(a, b, c, d, x[4], 7, -176418897);
        d = ff(d, a, b, c, x[5], 12, 1200080426);
        c = ff(c, d, a, b, x[6], 17, -1473231341);
        b = ff(b, c, d, a, x[7], 22, -45705983);
        a = ff(a, b, c, d, x[8], 7, 1770035416);
        d = ff(d, a, b, c, x[9], 12, -1958414417);
        c = ff(c, d, a, b, x[10], 17, -42063);
        b = ff(b, c, d, a, x[11], 22, -1990404162);
        a = ff(a, b, c, d, x[12], 7, 1804603682);
        d = ff(d, a, b, c, x[13], 12, -40341101);
        c = ff(c, d, a, b, x[14], 17, -1502002290);
        b = ff(b, c, d, a, x[15], 22, 1236535329);

        a = gg(a, b, c, d, x[1], 5, -165796510);
        d = gg(d, a, b, c, x[6], 9, -1069501632);
        c = gg(c, d, a, b, x[11], 14, 643717713);
        b = gg(b, c, d, a, x[0], 20, -373897302);
        a = gg(a, b, c, d, x[5], 5, -701558691);
        d = gg(d, a, b, c, x[10], 9, 38016083);
        c = gg(c, d, a, b, x[15], 14, -660478335);
        b = gg(b, c, d, a, x[4], 20, -405537848);
        a = gg(a, b, c, d, x[9], 5, 568446438);
        d = gg(d, a, b, c, x[14], 9, -1019803690);
        c = gg(c, d, a, b, x[3], 14, -187363961);
        b = gg(b, c, d, a, x[8], 20, 1163531501);
        a = gg(a, b, c, d, x[13], 5, -1444681467);
        d = gg(d, a, b, c, x[2], 9, -51403784);
        c = gg(c, d, a, b, x[7], 14, 1735328473);
        b = gg(b, c, d, a, x[12], 20, -1926607734);

        a = hh(a, b, c, d, x[5], 4, -378558);
        d = hh(d, a, b, c, x[8], 11, -2022574463);
        c = hh(c, d, a, b, x[11], 16, 1839030562);
        b = hh(b, c, d, a, x[14], 23, -35309556);
        a = hh(a, b, c, d, x[1], 4, -1530992060);
        d = hh(d, a, b, c, x[4], 11, 1272893353);
        c = hh(c, d, a, b, x[7], 16, -155497632);
        b = hh(b, c, d, a, x[10], 23, -1094730640);
        a = hh(a, b, c, d, x[13], 4, 681279174);
        d = hh(d, a, b, c, x[0], 11, -358537222);
        c = hh(c, d, a, b, x[3], 16, -722521979);
        b = hh(b, c, d, a, x[6], 23, 76029189);
        a = hh(a, b, c, d, x[9], 4, -640364487);
        d = hh(d, a, b, c, x[12], 11, -421815835);
        c = hh(c, d, a, b, x[15], 16, 530742520);
        b = hh(b, c, d, a, x[2], 23, -995338651);

        a = ii(a, b, c, d, x[0], 6, -198630844);
        d = ii(d, a, b, c, x[7], 10, 1126891415);
        c = ii(c, d, a, b, x[14], 15, -1416354905);
        b = ii(b, c, d, a, x[5], 21, -57434055);
        a = ii(a, b, c, d, x[12], 6, 1700485571);
        d = ii(d, a, b, c, x[3], 10, -1894986606);
        c = ii(c, d, a, b, x[10], 15, -1051523);
        b = ii(b, c, d, a, x[1], 21, -2054922799);
        a = ii(a, b, c, d, x[8], 6, 1873313359);
        d = ii(d, a, b, c, x[15], 10, -30611744);
        c = ii(c, d, a, b, x[6], 15, -1560198380);
        b = ii(b, c, d, a, x[13], 21, 1309151649);
        a = ii(a, b, c, d, x[4], 6, -145523070);
        d = ii(d, a, b, c, x[11], 10, -1120210379);
        c = ii(c, d, a, b, x[2], 15, 718787259);
        b = ii(b, c, d, a, x[9], 21, -343485551);

        this.a = add32(this.a, a);
        this.b = add32(this.b, b);
        this.c = add32(this.c, c);
        this.d = add32(this.d, d);
    }
}

const add32 = (a: number, b: number) => (a + b) | 0;

const rotateLeft = (value: number, bits: number) => (value << bits) | (value >>> (32 - bits));

const common = (q: number, a: number, b: number, x: number, s: number, t: number) => {
    return add32(rotateLeft(add32(add32(a, q), add32(x, t)), s), b);
};

const ff = (a: number, b: number, c: number, d: number, x: number, s: number, t: number) => {
    return common((b & c) | (~b & d), a, b, x, s, t);
};

const gg = (a: number, b: number, c: number, d: number, x: number, s: number, t: number) => {
    return common((b & d) | (c & ~d), a, b, x, s, t);
};

const hh = (a: number, b: number, c: number, d: number, x: number, s: number, t: number) => {
    return common(b ^ c ^ d, a, b, x, s, t);
};

const ii = (a: number, b: number, c: number, d: number, x: number, s: number, t: number) => {
    return common(c ^ (b | ~d), a, b, x, s, t);
};

const wordToHex = (word: number) => {
    let output = "";
    for (let i = 0; i < 4; i++) {
        output += `0${((word >>> (i * 8)) & 0xff).toString(16)}`.slice(-2);
    }
    return output;
};

const throwIfAborted = (signal?: AbortSignal) => {
    if (signal?.aborted) {
        throw new DOMException("上传已取消", "AbortError");
    }
};

const yieldToBrowser = () => new Promise<void>((resolve) => {
    if (typeof requestAnimationFrame === "function") {
        requestAnimationFrame(() => resolve());
        return;
    }
    setTimeout(resolve, 0);
});

const readBlobAsArrayBuffer = (blob: Blob, signal?: AbortSignal) => new Promise<ArrayBuffer>((resolve, reject) => {
    const reader = new FileReader();
    const cleanup = () => signal?.removeEventListener("abort", abort);
    const abort = () => {
        if (reader.readyState === FileReader.LOADING) {
            reader.abort();
            return;
        }
        cleanup();
        reject(new DOMException("上传已取消", "AbortError"));
    };
    if (signal?.aborted) {
        abort();
        return;
    }
    reader.onload = () => {
        cleanup();
        if (reader.result instanceof ArrayBuffer) {
            resolve(reader.result);
        } else {
            reject(new Error("读取上传文件失败"));
        }
    };
    reader.onerror = () => {
        cleanup();
        reject(reader.error || new Error("读取上传文件失败"));
    };
    reader.onabort = () => {
        cleanup();
        reject(new DOMException("上传已取消", "AbortError"));
    };
    signal?.addEventListener("abort", abort, {once: true});
    reader.readAsArrayBuffer(blob);
});

export const getFileMd5 = async (file: Blob, signal?: AbortSignal): Promise<string> => {
    const md5 = new Md5();
    let offset = 0;

    while (offset < file.size) {
        throwIfAborted(signal);
        const end = Math.min(offset + HASH_CHUNK_SIZE, file.size);
        const buffer = await readBlobAsArrayBuffer(file.slice(offset, end), signal);
        throwIfAborted(signal);
        md5.update(new Uint8Array(buffer));
        offset = end;
        await yieldToBrowser();
    }

    throwIfAborted(signal);
    return md5.digest();
};
