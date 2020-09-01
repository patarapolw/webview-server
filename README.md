# Practical web server in Golang with clean-up function

So, I tried to write a web server in Golang to fit in with [zserge/lorca](https://github.com/zserge/lorca). Focusing on [maximize/fullscreen on all platforms](https://github.com/webview/webview/issues/458) as well.

See [the original post](https://dev.to/patarapolw/practical-web-server-in-vanilla-go-with-clean-up-function-i-don-t-really-know-what-i-am-doing-1nh5).

Tested with cURL's

```sh
% PORT=3000 go run .
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

<small>I cannot upload <code>*.app</code> directly to GitHub Releases.</small>

[`darwin`](https://en.wikipedia.org/wiki/Darwin_%28operating_system%29) binaries can used for macOS, although not built natively on macOS.

## Open without Chrome

You can also open from Terminal with

```sh
./webview-server{DEPEND_ON_PLATFORM}
```

Provided that you can also revert to windowed mode with environmental variable `WINDOW=1`.
