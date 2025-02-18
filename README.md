# ROB - Refresh Only Builds

ROB is a proxy web server that rebuilds your application **only when a request comes in** after code 
changes, instead of triggering a rebuild on every file change. This approach reduces unnecessary builds and avoids refreshing multiple times
for your app to be ready.

## How It Works
1. **Starts a Reverse Proxy**: Forwards incoming requests to the actual application.
2. **Get notified of code changes**: Use [air](https://github.com/air-verse/air) or any other similar tool to call `rob --notify` when files change.
3. **Triggers a Rebuild**: If code has changed since the last request, it executes the specified build command on the next request, not on every change.
4. **Restarts the Application**: If the build succeeds, the app is restarted before forwarding the request.

## Features
 * Save battery life rebuilding only when you are ready to test your change.
 * Be sure you are seeing the latest code changes with a single refresh.
 * No need to refresh multiple times waiting for your app to be ready.
 * See build errors right in your browser when you refresh after a change.

<img width="1153" src="https://github.com/user-attachments/assets/f796c1a8-c0bf-47fc-8074-65cd9352ae39" />

### Limitations
 * Your `run` script is responsible for stopping and starting your app in the background.

Sample run script:

```sh
#!/bin/sh

pkill myapp
./myapp &
```

Sample build script:

```sh
#!/bin/sh

set -e
templ generate
go build -o myapp .
```

Sample air config:

```toml
[build]
  bin = ""
  cmd = "rob --notify"
  delay = 100
  full_bin = "true"
```

## Installation

```sh
go install github.com/tigrang/rob@latest
```

## Usage
Run ROB with the necessary flags:

```sh
rob --proxy localhost:3000
```

### Available Flags
| Flag            | Default Value            | Description                                             |
|-----------------|--------------------------|---------------------------------------------------------|
| `--notify`      | `false`                  | Notify the proxy to trigger a rebuild.                  |
| `--bin`         | `./run`                  | Path to the app startup script.                         |
| `--cmd`         | `./build`                | Command to execute for rebuilding the application.      |
| `--proxybind`   | `localhost:9000`         | Address where ROB listens for requests.                 |
| `--proxy`       | `localhost:3000`         | URL of the application to forward requests to.          |
| `--notifyroute` | `/internal/build/notify` | Path used to notify ROB of changes.                     |
| `--timeout`     | `30`                     | Time (seconds) to wait for the app to become available. |
| `--path`        | ``                       | Path to app                                             |

## License
[MIT License](LICENSE)


