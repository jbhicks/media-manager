# Media Manager

Go/Fyne desktop app for browsing, tagging, and previewing a local media library. Optional headless service for Jackett search, torrent downloads, and Jellyfin refresh.

See [ARCHITECTURE.md](ARCHITECTURE.md) for package layout and [SERVICE_QUICKSTART.md](SERVICE_QUICKSTART.md) for the download service.

**Stack:** Go 1.24, Fyne, SQLite (GORM), FFmpeg for thumbnails and hover previews.

## Desktop app

```console
make build    # ./tmp/media-manager.exe (Windows GUI flags)
make dev      # scripts/dev-server.sh
```

Requires [FFmpeg](https://ffmpeg.org/) on `PATH`. Thumbnails cache under `~/.media-manager/thumbnails/`.

Optional GPU: CUDA or VAAPI for video processing.

## Download service

```console
make build-service    # ./bin/media-manager-service
```

Optional integrations: Jackett, Transmission, Jellyfin, TMDb. Copy `.env.example` and fill API keys. Do not commit `.env`.

## Test

```console
make test    # go test ./...
```

There is not a meaningful test suite yet. Do not treat a green local `go test` as coverage.

## License

MIT. See [LICENSE](LICENSE).
