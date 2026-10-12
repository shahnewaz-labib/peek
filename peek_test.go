package peek

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"google.golang.org/genai"
)

// TestAskLocalFile runs the full flow against a fake Gemini API:
// upload -> wait for ACTIVE -> create interaction -> delete file.
func TestAskLocalFile(t *testing.T) {
	PollInterval = 0
	var (
		mu      sync.Mutex
		created map[string]any
		deleted bool
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/upload/v1beta/files":
			w.Header().Set("X-Goog-Upload-URL", "http://"+r.Host+"/upload-session")
		case r.URL.Path == "/upload-session":
			w.Header().Set("X-Goog-Upload-Status", "final")
			io.WriteString(w, `{"file":{"name":"files/abc","uri":"https://files/abc","mimeType":"video/mp4","state":"PROCESSING"}}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1beta/files/abc":
			io.WriteString(w, `{"name":"files/abc","uri":"https://files/abc","mimeType":"video/mp4","state":"ACTIVE"}`)
		case r.Method == http.MethodDelete && r.URL.Path == "/v1beta/files/abc":
			deleted = true
			io.WriteString(w, `{}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1beta/interactions":
			json.NewDecoder(r.Body).Decode(&created)
			io.WriteString(w, `{"id":"i1","status":"completed","steps":[{"type":"processing_call"},{"type":"model_output","content":[{"type":"text","text":"Error at 00:42"}]}]}`)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	video := filepath.Join(t.TempDir(), "clip.mp4")
	os.WriteFile(video, []byte("not really a video"), 0o600)

	ctx := context.Background()
	c, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:      "test",
		Backend:     genai.BackendGeminiAPI,
		HTTPOptions: genai.HTTPOptions{BaseURL: srv.URL},
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := Ask(ctx, c, video, "What error shows?", "")
	if err != nil {
		t.Fatal(err)
	}
	if got != "Error at 00:42" {
		t.Errorf("answer = %q", got)
	}
	if !deleted {
		t.Error("uploaded file was not deleted")
	}

	body, _ := json.Marshal(created)
	for _, want := range []string{`"model":"gemini-3.8-flash"`, `"processing":"agentic"`, `"uri":"https://files/abc"`, `"mime_type":"video/mp4"`, `"text":"What error shows?"`} {
		if !strings.Contains(string(body), want) {
			t.Errorf("request body missing %s\nbody: %s", want, body)
		}
	}
	// Background mode fails with File API URIs ("Unsupported file uri").
	if strings.Contains(string(body), `"background"`) {
		t.Errorf("request must not set background\nbody: %s", body)
	}
}
