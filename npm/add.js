import path from 'path'
import fs from 'fs'
import https from 'https'
import * as sys from './sys.js'

const ok = 200
const host = 'https://github.com/sugawarayuuta/nagi/releases/download/'
const os = sys.os()
const cpu = sys.cpu()
const dir = sys.dir()
const pkg = JSON.parse(await fs.promises.readFile(path.join(dir, 'package.json'), 'utf-8'))
const name = `/nagi-${os}-${cpu}${os === 'win' ? '.exe' : ''}`

function download(url) {
    const stream = fs.createWriteStream(path.join(dir, name))
    return new Promise((res, rej) => {
        const req = https.get(url, (msg) => {
            if (msg.headers.location) {
                download(msg.headers.location)
                    .then(() => res())
                    .catch((err) => rej(err))
            } else {
                if (msg.statusCode != ok) {
                    rej(`bad status code: ${msg.statusCode}`)
                }
                msg.pipe(stream)
            }
        })
        req.on('error', (err) => {
            rej(err)
        })
        stream.on('error', (err) => {
            rej(err)
        })
        stream.on('finish', () => {
            stream.close()
            res()
        })
    })
}

await download(host + pkg.version + name)
if (os != 'win') await fs.promises.chmod(path.join(dir, name), 0o755)