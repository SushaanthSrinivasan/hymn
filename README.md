# hymn

A lightweight TUI music player. Streams YouTube audio ad-free via mpv + yt-dlp,
with album art rendered inside the terminal.

## Requirements

`mpv` and `yt-dlp` must be on PATH (or pointed to via `config.toml`).

| Platform | Install hint |
|---|---|
| Windows | `winget install shinchiro.mpv` and `winget install yt-dlp.yt-dlp` |
| macOS   | `brew install mpv yt-dlp` |
| Linux   | `apt install mpv yt-dlp` (or your distro's equivalent) |

## Build

```
go build -ldflags="-s -w" -o hymn.exe .
```

The binary is fully self-contained.

## Run

```
./hymn
```

State (queue, history, playlists, cached audio) lives at:

- Windows: `%APPDATA%\hymn\`, audio cache under `%LOCALAPPDATA%\hymn\audio\`
- macOS / Linux: `~/.config/hymn/`, cache under `~/.cache/hymn/audio/`

## Keys

| Key | Action |
|---|---|
| `/` or `Ctrl+F` | open search |
| `enter` | search modal: enqueue + play; queue: play selected |
| `↑ ↓` / `j k` | move queue cursor |
| `space` | play / pause |
| `← →` / `h l` | seek ±5s |
| `+ -` | volume ±5 |
| `n` / `p` | next / prev (Ctrl+→ / Ctrl+←) |
| `a` | add cursor track to a playlist |
| `Shift+P` | open playlists modal |
| `q` / `Ctrl+C` | quit |

In the playlist modal: `Enter` appends the playlist to the queue, `Shift+Enter`
replaces the queue and starts playback.

## Mouse

Click queue rows to jump; click `[<<] [||] [>>]` to skip/pause; scroll to move
the queue cursor. Cell-motion mode is on by default.

## Album art

Detection order: Kitty graphics → iTerm2 inline images → Sixel → ANSI half-block
fallback (works in any 24-bit-color terminal, including Windows Terminal).

## Config

`config.toml` is auto-created on first run. See `internal/config/config.go` for
the full set of fields.

## Caching

Streamed audio is recorded to `<cache_dir>/<video_id>.cache` so that the second
play starts instantly. Total cache size is bounded by `cache.max_size_mb`
(default 2048 MB). LRU eviction runs at startup.

## Status

v0.1 — usable but minimal. Does not handle: livestreams, local files, lyrics,
scrobbling, MPRIS / SMTC, configurable keybinds, or theme switching.
