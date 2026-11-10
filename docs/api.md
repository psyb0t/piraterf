---
title: "HTTP and WebSocket API"
description: "The PIrateRF HTTP routes and WebSocket events the web UI uses: uploads, the file index, starting and stopping transmissions, files, playlists and presets."
nav_order: 9
---

# HTTP and WebSocket API

The web UI is a client of this API, so anything it does you can script. There is no authentication; read [Security](security.md) first.

## Contents

- [HTTP routes](#http-routes)
- [WebSocket events](#websocket-events)
- [Transmissions](#transmissions)
- [Files and playlists](#files-and-playlists)
- [Presets](#presets)

## HTTP routes

The router is in [`internal/pkg/services/piraterf/http_server.go`](https://github.com/psyb0t/piraterf/blob/main/internal/pkg/services/piraterf/http_server.go).

| Method and path | What it does |
| --- | --- |
| `GET /` | The web UI (`index.html` with cache-busting asset URLs). |
| `POST /upload` | Multipart upload. The file goes in the `file` field and the target mode in the `module` field, which decides the conversion: audio to WAV for FM, images for SSTV and Spectrum Paint, data for FSK, IQ for `sendiq`. Responds with JSON describing the stored file. |
| `GET /static/...` | Files from `PIRATERF_STATICDIR`. |
| `GET /files/...` | Files from `PIRATERF_FILESDIR`. A directory returns a JSON index of its entries. |
| `GET /ws` | The event WebSocket. |
| `GET /wsunix` | The live microphone bridge. It sends a `wsunixbridge.init` event with a `writerSocket` path, then forwards the audio the browser sends into that Unix socket for the `audiosock-broadcast` module. |

## WebSocket events

Every message in both directions is a JSON event:

```json
{
  "id": "uuid",
  "type": "rpitx.execution.start",
  "data": {},
  "timestamp": 1700000000,
  "metadata": {},
  "triggeredBy": null
}
```

`type` picks the handler and `data` carries its payload. Replies and status events go to **every** connected client, not only the sender, which is how several devices share one transmitter. `echo.request` gets an `echo.reply` and serves as the UI's heartbeat.

## Transmissions

Only one transmission runs at a time, because the GPIO pin cannot be shared.

| Event | Direction | Data |
| --- | --- | --- |
| `rpitx.execution.start` | client to server | `moduleName`, `args` (the module's settings object), `timeout` in seconds (0 means none), `playOnce`, optional `intro` and `outro` audio files |
| `rpitx.execution.started` | broadcast | `moduleName`, `args`, `initiatingClientId`, `timestamp` |
| `rpitx.execution.output-line` | broadcast | `type`, `line`, `timestamp`: one line of the module's output |
| `rpitx.execution.stop` | client to server | none |
| `rpitx.execution.stopped` | broadcast | `initiatingClientId`, `stoppingClientId`, `timestamp` |
| `rpitx.execution.error` | broadcast | `error`, `message`, `timestamp` |

`moduleName` is one of the gorpitx module names listed in [Transmission modes](transmission-modes.md). The shape of `args` is defined per module by [gorpitx](https://github.com/psyb0t/gorpitx).

## Files and playlists

| Event | Data | Replies |
| --- | --- | --- |
| `file.rename` | `filePath` (full path), `newName` (file name only) | `file.rename.success` or `file.rename.error` |
| `file.delete` | `filePath` (full path) | `file.delete.success` or `file.delete.error` |
| `audio.playlist.create` | `playlistFileName`, `files` (full paths) | `audio.playlist.create.success` with `playlistName` and `filePath`, or `audio.playlist.create.error` |

## Presets

Preset data is the settings object for one module, stored as described in [Presets](presets.md).

| Event | Data | Replies |
| --- | --- | --- |
| `preset.load` | `moduleName`, `presetName` | `preset.load.success` with `data`, or `preset.load.error` |
| `preset.save` | `moduleName`, `presetName`, `data` | `preset.save.success` or `preset.save.error` |
| `preset.rename` | `moduleName`, `presetName`, `newName` | `preset.rename.success` or `preset.rename.error` |
| `preset.delete` | `moduleName`, `presetName` | `preset.delete.success` or `preset.delete.error` |

Every error reply carries `error` and `message` strings and a `timestamp`.
