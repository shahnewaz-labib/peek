// Command peek asks Gemini a question about a video file or a YouTube URL.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/shahnewaz-labib/peek"
	"google.golang.org/genai"
)

func main() {
	model := flag.String("model", peek.DefaultModel, "Gemini model")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `usage: peek [-model m] <video-path|youtube-url> "<question>"

Needs GEMINI_API_KEY. A local file is uploaded to Google and deleted after the answer.`)
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}

	if os.Getenv("GEMINI_API_KEY") == "" && os.Getenv("GOOGLE_API_KEY") == "" {
		fmt.Fprintln(os.Stderr, "peek: GEMINI_API_KEY is not set (get one at https://aistudio.google.com/apikey)")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// No client timeout: a long video can take minutes. Ctrl-C cancels ctx.
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend:    genai.BackendGeminiAPI,
		HTTPClient: &http.Client{},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "peek:", err)
		os.Exit(1)
	}
	answer, err := peek.Ask(ctx, client, flag.Arg(0), flag.Arg(1), *model)
	if err != nil {
		fmt.Fprintln(os.Stderr, "peek:", err)
		os.Exit(1)
	}
	fmt.Println(answer)
}
