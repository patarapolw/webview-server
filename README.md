# Practical web server in Golang with clean-up function

So, I tried to write a web server in Golang to fit in with [zserge/lorca](https://github.com/zserge/lorca). Focusing on [maximize/fullscreen on all platforms](https://github.com/webview/webview/issues/458) as well.

See [the original post](https://dev.to/patarapolw/practical-web-server-in-vanilla-go-with-clean-up-function-i-don-t-really-know-what-i-am-doing-1nh5).

Tested with cURL's

```sh
% PORT=3000 ./webview-server
% curl -i -X PUT --data 'hello' http://127.0.0.1:3000/api/file\?filename\=test.txt
```

This, by default, works with [lokijs](https://github.com/techfort/LokiJS), by using a custom adaptor.

```ts
import Loki from 'lokijs'

class LokiRestAdaptor {
  loadDatabase (dbname: string, callback: (data: string | null | Error) => void) {
    fetch(`/api/file?filename=${encodeURIComponent(dbname)}`)
      .then((r) => r.text())
      .then((r) => callback(r))
      .catch((e) => callback(e))
  }

  saveDatabase (dbname: string, dbstring: string, callback: (e: Error | null) => void) {
    fetch(`/api/file?filename=${encodeURIComponent(dbname)}`, {
      method: 'PUT',
      body: dbstring
    })
      .then(() => callback(null))
      .catch((e) => callback(e))
  }

  deleteDatabase (dbname: string, callback: (data: Error | null) => void) {
    fetch(`/api/file?filename=${encodeURIComponent(dbname)}`, {
      method: 'DELETE'
    })
      .then(() => callback(null))
      .catch((e) => callback(e))
  }
}

// eslint-disable-next-line import/no-mutable-exports
export let loki: Loki

export async function initDatabase () {
  return new Promise((resolve) => {
    loki = new Loki('db.loki', {
      adapter: new LokiRestAdaptor(),
      autoload: true,
      autoloadCallback: () => {
        resolve()
      },
      autosave: true,
      autosaveInterval: 4000
    })
  })
}

window.onbeforeunload = (e: Event) => {
  if (loki) {
    if (loki.autosaveDirty()) {
      loki.saveDatabase()

      e.preventDefault()
      e.returnValue = false
    }
  }
}
```

And, in you Webpack dev server (e.g. `webpack.config.js`)

```js
// @ts-check

/* eslint-disable @typescript-eslint/no-var-requires */
const fs = require('fs')
const path = require('path')

const morgan = require('morgan')

const BINARY_DIR = 'release'

module.exports = {
  devServer: {
    /**
     *
     * @param {import('express').Express} app
     * @param {import('http').Server} server
     * @param {*} compiler
     */
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    before (app, server, compiler) {
      app.use('/api', morgan('tiny'))

      app.get('/api/file', (req, res) => {
        /** @type {*} */
        const { filename } = req.query
        const p = path.resolve(__dirname, BINARY_DIR, filename)

        if (fs.existsSync(p)) {
          fs.createReadStream(p).pipe(res)
          return
        }

        res.sendStatus(404)
      })

      app.put('/api/file', (req, res) => {
        /** @type {*} */
        const { filename } = req.query
        const p = path.resolve(__dirname, BINARY_DIR, filename)

        req.pipe(fs.createWriteStream(p))
        res.sendStatus(201)
      })

      app.delete('/api/file', (req, res) => {
        /** @type {*} */
        const { filename } = req.query
        const p = path.resolve(__dirname, BINARY_DIR, filename)

        fs.unlinkSync(p)
        res.sendStatus(201)
      })
    }
  }
}
```

## Web browser in use

Currently, this app doesn't bundle a web browser. Instead, it uses Chrome DevTools Protocol; therefore, either Chrome or Chromium must be installed.

See [/deps.md](/deps.md).

## Security concerns

I learnt this from [pywebview](https://pywebview.flowrl.com/guide/security.html). A major thing about this, is [CSRF attack](https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)).

## Customization

Please see [/config/types.go](/config/types.go). The easiest way is to create `/config.yaml` alongside the built `webview-server*`.

## Building

You can also build for your platform, or multiple platforms at once -- take a peek inside [robo.yml](/robo.yml)

Note that executables in macOS can also run in windowed mode (no console), by renaming the extension to `*.app`. No need to make a folder of `*.app/`.

I cannot upload `*.app` directly to GitHub Releases.

[`darwin`](https://en.wikipedia.org/wiki/Darwin_%28operating_system%29) binaries can used for macOS, although not built natively on macOS.
