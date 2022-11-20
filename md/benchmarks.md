### Benchmark details

· tool: hyperfine

· cache/lock: completely cold

· command: `hyperfine --runs 10  --prepare ../clean.sh --show-output 'nagi add express' 'bun add express' 'pnpm add express' 'yarn add express' 'npm add express'`

· ../clean.sh: [clean.sh](../clean.sh)

· CPU: Quad-Core Intel Core i5

· architecture: x64

· OS: MacOS