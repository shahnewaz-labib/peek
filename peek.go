// Package peek asks Gemini a question about a video and returns the text answer.
package peek

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"google.golang.org/genai"
	ia "google.golang.org/genai/interactions/models/interactions"
	"google.golang.org/genai/interactions/models/operations"
)

// DefaultModel supports agentic video processing.
const DefaultModel = "gemini-3.8-flash"

// PollInterval is how long Ask waits between upload status checks.
var PollInterval = 5 * time.Second

var urlRe = regexp.MustCompile(`^https?://`)

// Ask sends src (a local video path or a public YouTube URL) and question to
// model, and returns the answer. A local file is uploaded through the File API
// and deleted before Ask returns. An empty model means DefaultModel.
//
// The answer comes back in one long request. The SDK's default HTTP client
// times out after 60s, so give c an HTTPClient without a short timeout and
// use ctx for cancellation.
func Ask(ctx context.Context, c *genai.Client, src, question, model string) (string, error) {
	if model == "" {
		model = DefaultModel
	}
	video := ia.VideoContent{Processing: genai.Ptr(ia.NewProcessing(ia.ProcessingEnumAgentic))}

	if urlRe.MatchString(src) {
		video.URI = &src
	} else {
		f, err := upload(ctx, c, src)
		if err != nil {
			return "", err
		}
		defer c.Files.Delete(context.WithoutCancel(ctx), f.Name, nil)
		video.URI = &f.URI
		video.MimeType = genai.Ptr(ia.VideoContentMimeType(f.MIMEType))
	}

	input := ia.NewInteractionsInput([]ia.Content{
		ia.NewContent(video),
		ia.NewContent(ia.TextContent{Text: question}),
	})
	// Synchronous on purpose: with background mode, polling fails with
	// "Unsupported file uri" for File API uploads.
	res, err := c.Interactions.Create(ctx, operations.CreateInteractionRequest{
		Body: operations.NewCreateInteractionRequestBody(ia.CreateModelInteraction{
			Model: ia.Model(model),
			Input: &input,
		}),
	})
	if err != nil {
		return "", fmt.Errorf("create interaction: %w", err)
	}
	it := res.Interaction
	if it == nil || it.ID == nil {
		return "", fmt.Errorf("create interaction: empty response")
	}
	if it.Status != ia.InteractionStatusCompleted {
		return "", fmt.Errorf("interaction %s: status %s", *it.ID, it.Status)
	}
	if answer := text(it); answer != "" {
		return answer, nil
	}
	return "", fmt.Errorf("interaction %s: no text in response", *it.ID)
}

func upload(ctx context.Context, c *genai.Client, path string) (*genai.File, error) {
	f, err := c.Files.UploadFromPath(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("upload %s: %w", path, err)
	}
	for f.State == genai.FileStateProcessing {
		if err := sleep(ctx); err != nil {
			return nil, err
		}
		if f, err = c.Files.Get(ctx, f.Name, nil); err != nil {
			return nil, fmt.Errorf("get file: %w", err)
		}
	}
	if f.State != genai.FileStateActive {
		c.Files.Delete(context.WithoutCancel(ctx), f.Name, nil)
		return nil, fmt.Errorf("file %s: state %s", f.Name, f.State)
	}
	return f, nil
}

func text(it *ia.Interaction) string {
	if it.OutputText != nil {
		return *it.OutputText
	}
	var b strings.Builder
	for _, s := range it.GetSteps() {
		if s.ModelOutputStep == nil {
			continue
		}
		for _, c := range s.ModelOutputStep.GetContent() {
			if c.TextContent != nil {
				b.WriteString(c.TextContent.Text)
			}
		}
	}
	return b.String()
}

func sleep(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(PollInterval):
		return nil
	}
}
