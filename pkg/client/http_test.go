package client

import "testing"

func TestSanitizeValidURL(t *testing.T) {
	urls := []struct {
		Input, Output string
	}{
		{"http://localhost:23110", ""},
		{"https://localhost:23110", ""},
		{"https://localhost", "https://localhost:23110"},
		{"http://localhost", "http://localhost:23110"},
		{"http://:23110", ""},
		{"http://127.0.0.1", "http://127.0.0.1:23110"},
		{"https://127.0.0.1:23110", ""},
		{"127.0.0.1:23110", "http://127.0.0.1:23110"},
		{"localhost", "http://localhost:23110"},
		{":23110", "http://:23110"},
		{"http://:23110", ""},
	}
	for idx, url := range urls {
		out, err := sanitizeURL(url.Input)
		if err != nil {
			t.Errorf("#%d: %q is invalid, while expected it to be valid: %v", idx, url.Input, err)
			continue
		}
		if url.Output == "" {
			url.Output = url.Input
		}
		if url.Output != out {
			t.Errorf("#%d: %q != %q", idx, url.Output, out)
		}
	}
}

func TestSanitizeInvalidURL(t *testing.T) {
	urls := []struct {
		Input string
	}{
		{""},
	}
	for idx, url := range urls {
		_, err := sanitizeURL(url.Input)
		if err == nil {
			t.Errorf("#%d: %q is valid, while expected it to be invalid", idx, url.Input)
		}
	}
}
