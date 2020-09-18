# Cloak

A simple GUI app for gocryptfs, using web tech stack.
This is currently a work in progress, it works on macOS and Linux, but you have to build it from source.

UI / interaction mimicked from Cryptomator.

# Why

I wrote a similar GUI called [Cloaklet](https://github.com/Cloaklet/Cloaklet) using QML + Golang.
However I don't enjoy developing in QML at all, and it had some issues I failed to resolve, so I quickly gave it up.
After that I went back to use Cryptomator, but its UI feels slow and somehow inconsistent.
This time I got an idea from one of my early projects to use web browser as UI renderer, thus this new project.

# To build

## For macOS

Notice: you have to use a running instance of macOS, either a VM or a real Apple computer.

- Install Xcode related stuff with `xcode-select --install`.
- Run `go run build.go build` and it should create the `Cloak.app` bundle.
- Double click to start the app.

The macOS app bundle includes `gocryptfs` binary, so you don't to compile and install it manually.

## For Linux

- Install required libraries: `sudo apt install libappindicator3-dev gcc libgtk-3-dev`.
- If you're on Linux Mint, install an additional library: `sudo apt install libxapp-dev`.
- Run `go run build.go build` and it should produce `cloak` binary.

Please note that we don't currently create app bundle for Linux, so `gocryptfs` is not included.
Drop an official binary from [gocryptfs release](https://github.com/rfjakob/gocryptfs/releases) into anywhere of your `PATH` should work.

# To develop

## Frontend

The frontend (UI) project resides in `frontend` directory. It's a standard Vue project managed by vue-cli.
Simply run `npm run serve` inside `frontend` directory.
You can also run the `serve` task from vue-cli UI, run `vue ui` to get started.

## Backend

You should build the frontend project first so the backend can find assets for the UI.

- Inside the `frontend` directory, run `npm run build`.
- To run the app, simply invoke `go run .` in the project root.

# Notice

- `gocryptfs` requires `FUSE` to function. For macOS please install `OSXFUSE`.
- Windows is not supported, because `gocryptfs` does not work on Windows.
- Avoid committing `statik` module because it contains large blob of files produced by the frontend project.

# License

GPL v3, see LICENSE file.