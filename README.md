# Cloak

[![Build Status](https://img.shields.io/travis/com/Cloaklet/Cloak?style=for-the-badge)](https://travis-ci.com/Cloaklet/Cloak)

A simple GUI app for gocryptfs, using web tech stack. Works on macOS and Linux.

UI / interaction mimicked from Cryptomator. English and Simplified Chinese are included.

# Usage

- Download prebuilt binaries from [releases page](https://github.com/Cloaklet/Cloak/releases/latest/).
- For Linux users, set executable permission for the `AppImage` file (alternatively you can use `chmod +x Cloak*.AppImage`), then just run it.
- For macOS users, after decompressing the ZIP archive, you might need to run `xattr -d -r com.apple.quarantine Cloak.app` in Terminal, otherwise GateKeeper would refuse to run the app.
- You can open the UI or quit the app via `Open` menu item of the tray icon (or menubar icon).

# Why

I wrote a similar GUI called [Cloaklet](https://github.com/Cloaklet/Cloaklet) using QML + Golang.
However, I don't enjoy developing in QML at all, and quickly ran into some issues which I can't resolve.
After that I went back to use Cryptomator, but its UI feels slow and somehow inconsistent.
This time I got an idea from one of my early projects to use web browser as UI renderer, thus this new project.

# To build

## For macOS

Notice: you have to use a running instance of macOS, either a VM or a real Apple computer.

- Install Xcode related stuff with `xcode-select --install`.
- Install frontend dependencies: `cd frontend && npm install`.
- Run `go run build.go build` in project root, and it should create the `Cloak.app` bundle.
- Double click to start the app.

## For Linux

- Install required libraries: `sudo apt install libappindicator3-dev gcc libgtk-3-dev libxapp-dev`.
- Install frontend dependencies: `cd frontend && npm install`.
- Run `go run build.go build` and it should produce an AppImage binary.

The AppImage binary includes all required libraries and tools, so you can run it right away.

# To develop

## Frontend

The frontend (UI) project resides in `frontend` directory. It's a standard Vue project managed by vue-cli.

- Install dependencies: `npm install` inside `frontend` directory.
- Simply run `npm run serve` inside `frontend` directory.
- You can also run the `serve` task from vue-cli UI, run `vue ui` to get started.

## Backend

You should build the frontend project first so the backend can find assets for the UI.

- Inside the `frontend` directory, run `npm run build`.
- To run the app, simply invoke `go run .` in the project root.

# Notice

- `gocryptfs` requires `FUSE` to function. For macOS please install `OSXFUSE`.
- Windows is not supported, because `gocryptfs` does not work on Windows.
- Avoid committing `statik` module because it contains large blob of files produced by the frontend project.
- If you are building the app yourself, missing `libxapp-dev` would not result in error; But when running the built AppImage on Linux Mint, menu item will lose highlighting.

# Credits

- [RemixIcon](https://remixicon.com/)
- [CSS-Spinner](https://github.com/loadingio/css-spinner/)

# License

GPL v3, see LICENSE file.