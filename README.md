# Practical web server in vanilla Go with clean-up function

So, I tried to write a web server in Golang to fit in with [webview/webview](https://github.com/webview/webview). Focusing on [maximize/fullscreen on all platforms](https://github.com/webview/webview/issues/458) as well.

See [the original post](https://dev.to/patarapolw/practical-web-server-in-vanilla-go-with-clean-up-function-i-don-t-really-know-what-i-am-doing-1nh5).

Tested with cURL's

```sh
% PORT=3000 go run .
% curl -i -X PUT --data 'hello' http://127.0.0.1:3000/api/file\?filename\=test.txt
```

I have another nice way to connect with frontend. You can argue that sending SQL and its parameters might be a better way...

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

## Security concerns

I learnt this from [pywebview](https://pywebview.flowrl.com/guide/security.html). A major thing about this is [CSRF attack](https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)).

Therefore, I plan to implement CSRF protection as well.

## Customization

Please see [custom.go](/custom.go). The easiest way is to create `config.json` alongside the built `webview-server[.exe]`.
