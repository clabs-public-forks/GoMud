# Getting Running on Various Platforms

These guides assume you want to build from the source. You can also download the latest package and run the appropriate pre-compiled binary for your platform from the [releases section](https://github.com/GoMudEngine/GoMud/releases) of the repo.

- [Raspberry PI Zero 2W](RASPBERRY-PI.md)
- [Running via Docker](DOCKER.md)
- [Setting Up an EC2 Instance](EC2.md)


# Quick Start

You can download the latest release from the [releases page](https://github.com/GoMudEngine/GoMud/releases), unzip it and run the binary to get started, or if you prefer to build it yourself, follow the instructions below.

A youtube playlist to getting started has been set up here:

[![Getting Started Videos](https://i.ytimg.com/vi/OOZqX01aHt8/hqdefault.jpg "Getting Started Playlist")](https://www.youtube.com/watch?v=OOZqX01aHt8&list=PL20JEmG_bxBuaOE9oFziAhAmx1pyXhQ1p)

You can compile and run it locally with:

> `go run .`

Or you can just build the binary if you prefer:

> `go build -o GoMudServer`

> `./GoMudServer`

Or if you have docker installed:

> `docker compose up --build`


# Automatic HTTPS

For simple single-server installs, GoMud can automatically manage Let's Encrypt certificates for the built-in web client and admin UI.

If you want a guided interactive setup, run:

> `make https-setup`

The helper updates `_datafiles/config.yaml` and creates a timestamped backup first.

1. Point a public DNS name at your server.
2. In `_datafiles/config.yaml`, set `FilePaths.WebDomain` to that hostname.
3. Set `Network.HttpPort` to `80` and `Network.HttpsPort` to `443`.
4. Optionally set `FilePaths.HttpsEmail` so Let's Encrypt can send expiry notices.
5. Leave `FilePaths.HttpsCertFile` and `FilePaths.HttpsKeyFile` empty unless you want to use your own certificate files instead.

Notes:

- Automatic HTTPS is intended for one public server that owns ports `80` and `443`.
- `localhost`, private-only names, and raw IP addresses will stay on HTTP.
- If automatic HTTPS cannot succeed, GoMud falls back to HTTP and logs what to fix.
- The certificate cache is stored in `FilePaths.HttpsCacheDir`, which defaults to `_datafiles/tls`.
- The admin page at `/admin/https/` shows the current HTTPS mode, checks, and recommended fixes.
