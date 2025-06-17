# vidlogd

### A Terminal-Based YouTube Video Logger

## Demo

https://github.com/user-attachments/assets/caaa27a4-07c2-4fc4-bc71-9f4496edd108

### Built With

![Go Badge](https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=fff&style=for-the-badge)

## Features

### Video Logging

- **YouTube Integration** - Automatically fetch video details from URLs
- **Rating System** - Rate videos with stars (0-5)
- **Review Notes** - Add your own thoughts and reviews
- **Data Management** - Edit, delete, and search through your video collection

### Analytics Dashboard

- **Comprehensive Stats** - Dashboard cards showing total videos, average rating, rewatch percentage, and channel count
- **Interactive Charts** - Visual representations of rating distribution and monthly activity trends
- **Channel Analytics** - Channel-specific statistics with average ratings and video counts
- **Search & Filter** - Fuzzy find videos by title and channel

## Prerequisites

- **Go 1.24+** - [Download here](https://golang.org/dl/)
- **YouTube Data API v3 Key** - [Get one here](https://developers.google.com/youtube/v3/getting-started)
- **Nerd Fonts** (recommended) - For proper Unicode symbol display

## Installation

### Option 1: Go Install (Recommended)

```bash
go install github.com/mamuzad/vidlogd@latest
```

### Option 2: Build from Source

```bash
git clone https://github.com/mamuzad/vidlogd.git
cd vidlogd
make build
./bin/vidlogd
```

## Quick Start

### 1. Configure YouTube API Key

You can set your YouTube API key in several ways:

**Option A: Environment Variable**

```bash
export YOUTUBE_API_KEY="********"
```

**Option B: .env File (in project directory)**

```bash
echo "YOUTUBE_API_KEY=********" > .env
```

**Option C: Through the App Settings**

- Enter your YouTube API key in the app settings

### 2. Run

```bash
vidlogd
```

## Todo

- [x] Settings view
- [ ] Build/config
  - [ ] Package/CLI
- [x] Search and filter videos
- [ ] Video thumbnails
  - [ ] Display image with protocols for kitty, wezterm, ghostty
- [x] Statistics view
- [x] Add short videos list to stats view
