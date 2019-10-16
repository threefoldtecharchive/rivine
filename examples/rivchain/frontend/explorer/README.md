# Explorer

A block explorer for rivchain

## Run it yourself

### Prerequisites
* Caddyserver
* rivchain daemon (`rivined`)


Make sure you have `rivined` (the rivchain daemon) running with the explorer module enabled:
`rivined -M cgte`

Now start caddy from the `caddy` folder of this repository:
`caddy -conf Caddyfile.local`
and browse to http://localhost:2015
