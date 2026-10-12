# peek

Ask Gemini a question about a video. Works with a local file or a public YouTube URL.

```
peek clip.mp4 "List each error message on screen, with its mm:ss timestamp."
```

```
local file ──upload (File API)──> wait until ACTIVE ──> ask (agentic mode) ──> print answer
                                                                        └──> delete the upload
YouTube URL ─────────────────────────────────────────> ask (agentic mode) ──> print answer
```

## Install

Download a binary from [Releases](https://github.com/shahnewaz-labib/peek/releases), or:

```
go install github.com/shahnewaz-labib/peek/cmd/peek@latest
export GEMINI_API_KEY=...   # https://aistudio.google.com/apikey
```

## Release

```
git tag v0.2.0 && git push origin v0.2.0
```

The `release` workflow runs the tests, then GoReleaser builds macOS, Linux, and Windows binaries and attaches them to a GitHub Release.

## Use

```
peek [-model gemini-3.8-flash] <video-path|youtube-url> "<question>"
```

- A local file goes to Google. peek deletes it after the answer. Google deletes unused uploads after 48 hours.
- Ask a specific question. Ask for timestamps.
- Exit code 0 means success, 1 means an error, 2 means bad arguments.

## Use as a library

```go
client, _ := genai.NewClient(ctx, &genai.ClientConfig{
	Backend:    genai.BackendGeminiAPI,
	HTTPClient: &http.Client{}, // the SDK default times out after 60s
})
answer, err := peek.Ask(ctx, client, "clip.mp4", "What happens at 0:42?", "")
```

`Ask` makes one synchronous request. Background mode fails with uploaded files (`Unsupported file uri`), so peek does not use it.

## Test

```
go test ./...
```

The test runs the full flow against a fake Gemini API. It does not call Google.
