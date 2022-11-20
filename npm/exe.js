#!/usr/bin/env node

import path from 'path'
import child from 'child_process'
import * as sys from './sys.js'

const os = sys.os()
const cpu = sys.cpu()
const dir = sys.dir()
const name = `/nagi-${os}-${cpu}${os === 'win' ? '.exe' : ''}`

child.execFileSync(path.join(dir, name), process.argv.slice(2), { stdio: 'inherit' })