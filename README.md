# Musa

<p align="center">
  <img src="assets/icon.png" alt="Musa icon" width="160" height="160">
</p>

![Musa presentation](assets/musa-presentation.gif)

A controller-friendly music player for your own library.

Musa is built with Go + Raylib and is designed around browsing albums, enjoying cover art, and moving through your music with a clean, console-like feel.

## Install

### macOS

Clone and install:

```bash
git clone https://github.com/katistix/musa.git && cd musa && go install ./cmd/musa-install && $(go env GOPATH)/bin/musa-install
```

This creates a standalone app bundle at:

```text
~/Applications/Musa.app
```

Then launch it with Finder, Spotlight, or:

```bash
open ~/Applications/Musa.app
```

## Run in development

```bash
go run .
```

## Features

- scans `~/Music`
- album-first browsing
- search and cover art
- controller-friendly navigation
- Now Playing view with waveform
- remembers last album, track, and playback position
- restores paused where you left off

## Notes

- On macOS, the installer creates a real `.app` bundle instead of leaving you with a console executable.
- Musa stores its playback state in:

```text
~/.musa-state.json
```
