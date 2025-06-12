# vidlogd

### A Terminal-Based YouTube Video Logger

## Demo

https://github.com/user-attachments/assets/caaa27a4-07c2-4fc4-bc71-9f4496edd108

### Built With

![Go Badge](https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=fff&style=for-the-badge)

## Features

- **YouTube Integration** - Automatically fetch video details from URLs
- **Rating System** - Rate videos with stars (0-5)
- **Review Notes** - Add your own thoughts and comments
- **Video History** - View all logged videos in a sortable table
- **Data Management** - Edit or delete entries as needed

## Installation

1. Get a YouTube API key and add it to `.env`:

   ```bash
   echo "YOUTUBE_API_KEY=********" > .env
   ```

2. Run:
   ```bash
   go mod tidy
   go run .
   ```

## Todo

- [x] Settings view
- [ ] Build/config
  - [ ] Package/CLI
- [ ] Search and filter videos
- [ ] Video thumbnails
  - [ ] Display image with protocols for kitty, wezterm, ghostty
- [ ] Export data (CSV/JSON)
- [ ] Tags/categories
- [ ] Import existing data
- [ ] Statistics view
