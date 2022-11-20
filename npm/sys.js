import url from 'url'
import path from 'path'

function os() {
    switch (process.platform) {
        case 'win32':
            return 'win'
        case 'darwin':
            return 'mac'
        case 'linux':
            return 'linux'
        default:
            throw new Error('unsupported os')
    }
}

function cpu() {
    switch (process.arch) {
        case 'arm':
            return 'arm32'
        case 'arm64':
            return 'arm64'
        case 'ia32':
            return '32'
        case 'x64':
            return '64'
        default:
            throw new Error('unsupported cpu')
    }
}

function dir() {
    return path.dirname(url.fileURLToPath(import.meta.url))
}

export { cpu, os, dir }