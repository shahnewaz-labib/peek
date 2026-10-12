---
name: ask-about-video
description: Use when a task needs the content of a video file (.mp4, .mov, .webm, screen recording, bug repro clip) or a YouTube URL, since the agent cannot open video itself.
---

# Ask About Video

The agent cannot watch video. Gemini can. Send Gemini the video plus one focused question with `peek`, and read its text answer.

## Steps

1. **Ask before upload.** peek uploads a local file to Google. Get the user's go-ahead for each file. A public YouTube URL needs no upload.
2. Run:
   ```bash
   peek <path-or-youtube-url> "<question>"
   ```
3. Ask for timestamps in the question, for example: "List each error message on screen, with its mm:ss timestamp."
4. **Check details that matter.** Gemini can be wrong. For a fact you will act on (an error text, a value, a UI state), extract that frame and look at it yourself:
   ```bash
   ffmpeg -loglevel error -ss 00:42 -i video.mp4 -frames:v 1 /tmp/frame-0042.png
   ```
   Then read the PNG.

## Quick reference

| Need | How |
|---|---|
| Different model | `peek -model <name> ...` |
| `peek: command not found` | `go install github.com/shahnewaz-labib/peek/cmd/peek@latest` |
| Missing API key | Ask the user to set `GEMINI_API_KEY` |
| Exact pixels at a moment | `ffmpeg` frame, then read it |

## Common mistakes

- A vague question ("what happens?") gives a vague summary. Ask for the exact thing you need.
- Treating Gemini's answer as ground truth for a fix. Check the frame.
- Uploading a private video without asking.
