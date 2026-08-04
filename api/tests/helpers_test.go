package tests

import (
	"crypto/sha256"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DanielJohn17/url-shortner/api/internal/helpers"
	"github.com/gin-gonic/gin"
)

func TestCanonicalizeBasic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     string
		wantErr  bool
	}{
		{name: "lowercases scheme and host", input: "HTTPS://Example.COM/Path", want: "https://example.com/Path"},
		{name: "removes default http port", input: "http://example.com:80/x", want: "http://example.com/x"},
		{name: "removes default https port", input: "https://example.com:443/x", want: "https://example.com/x"},
		{name: "keeps non-default port", input: "http://example.com:8080/x", want: "http://example.com:8080/x"},
		{name: "preserves query and fragment", input: "https://example.com/path?q=1#frag", want: "https://example.com/path?q=1#frag"},
		{name: "escapes space in path", input: "https://example.com/a b", want: "https://example.com/a%20b"},
		{name: "keeps trailing slash", input: "https://example.com/", want: "https://example.com/"},
		{name: "empty string", input: "", want: ""},
		{name: "no scheme", input: "example.com/path", want: "example.com/path"},
		{name: "rejects malformed percent escape", input: "https://example.com/%zz", wantErr: true},
		{name: "rejects missing protocol", input: "://bad", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := helpers.CanonicalizeBasic(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (result %q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func shortCode(t *testing.T, full string) (string, bool) {
	t.Helper()
	idx := strings.LastIndex(full, "/")
	if idx < 0 || idx == len(full)-1 {
		return "", false
	}
	return full[idx+1:], true
}

func TestHashString(t *testing.T) {
	empty := helpers.HashString("")
	want := sha256.Sum256([]byte(""))
	if empty != want {
		t.Errorf("empty hash mismatch: got %x, want %x", empty, want)
	}

	if len(helpers.HashString("x")) != 32 {
		t.Errorf("expected 32-byte hash, got %d", len(helpers.HashString("x")))
	}

	if helpers.HashString("hello") != helpers.HashString("hello") {
		t.Errorf("hash should be deterministic")
	}

	if helpers.HashString("hello") == helpers.HashString("world") {
		t.Errorf("different inputs should produce different hashes")
	}
}

func TestEncodeBytes(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{name: "empty slice", input: []byte{}, want: "0"},
		{name: "all zeros", input: make([]byte, 32), want: "0"},
		{name: "known value 255", input: []byte{255}, want: "47"},
		{name: "single byte one", input: []byte{1}, want: "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := helpers.EncodeBytes(tt.input); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}

	// Determinism
	if helpers.EncodeBytes([]byte("abc")) != helpers.EncodeBytes([]byte("abc")) {
		t.Errorf("EncodeBytes should be deterministic")
	}

	// Output should only contain base62 alphabet
	out := helpers.EncodeBytes([]byte("hello"))
	for _, r := range out {
		if !strings.ContainsRune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", r) {
			t.Errorf("output %q contains non-base62 char %q", out, r)
		}
	}
}

type urlPayload struct {
	LongURL string `json:"long_url" validate:"required,url"`
}

func TestParseJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid payload", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/", jsonBody(`{"long_url":"https://example.com/x"}`))
		var p urlPayload
		if err := helpers.ParseJSON(c, &p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.LongURL != "https://example.com/x" {
			t.Errorf("got %q", p.LongURL)
		}
	})

	t.Run("missing required field", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/", jsonBody(`{}`))
		var p urlPayload
		if err := helpers.ParseJSON(c, &p); err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("invalid url field", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/", jsonBody(`{"long_url":"not-a-url"}`))
		var p urlPayload
		if err := helpers.ParseJSON(c, &p); err == nil {
			t.Fatal("expected validation error for invalid url, got nil")
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/", jsonBody(`{"long_url":`))
		var p urlPayload
		if err := helpers.ParseJSON(c, &p); err == nil {
			t.Fatal("expected decode error, got nil")
		}
	})

	t.Run("empty body", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/", jsonBody(""))
		var p urlPayload
		if err := helpers.ParseJSON(c, &p); err == nil {
			t.Fatal("expected error for empty body, got nil")
		}
	})

	t.Run("nil body", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/", jsonBody(""))
		c.Request.Body = nil
		var p urlPayload
		err := helpers.ParseJSON(c, &p)
		if err == nil {
			t.Fatal("expected error for nil body, got nil")
		}
		if !strings.Contains(err.Error(), "Missing request body") {
			t.Errorf("got error %q", err.Error())
		}
	})

	t.Run("unknown extra fields ignored", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/", jsonBody(`{"long_url":"https://example.com/x","extra":"ignored"}`))
		var p urlPayload
		if err := helpers.ParseJSON(c, &p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
