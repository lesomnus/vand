package vand

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lesomnus/vand/cmd/config"
)

func ptr[T any](v T) *T { return &v }

func TestServer(t *testing.T) {
	s := NewServer(config.PackageConfig{
		// Root package: go.mod at the repository root.
		"": {Target: "https://github.com/Holiday-Robot/hday-go"},
		// Ordinary package: go.mod at the repository root.
		"x/net": {Target: "https://github.com/golang/net", Branch: "master"},
		// Subdirectory module: go.mod lives in pretty/ of the same repo,
		// so the advertised prefix is the domain root.
		"pretty": {Target: "https://github.com/Holiday-Robot/hday-go", Root: ptr("")},
		// Nested subdirectory module: advertised prefix is an ancestor path.
		"svc/kamino": {Target: "https://github.com/Holiday-Robot/svc", Root: ptr("svc")},
		// Hidden package.
		"secret": {Target: "https://github.com/Holiday-Robot/secret", Hidden: true},
	})

	tests := []struct {
		name    string
		target  string // request URL
		goGet   bool
		code    int
		content string // expected go-import content, "" when 404 expected
		noindex bool
	}{
		{
			name:    "root package resolves to host without trailing slash",
			target:  "http://hday.dev/",
			goGet:   true,
			code:    200,
			content: "hday.dev git https://github.com/Holiday-Robot/hday-go",
		},
		{
			name:    "ordinary package advertises its full path",
			target:  "http://hday.dev/x/net",
			goGet:   true,
			code:    200,
			content: "hday.dev/x/net git https://github.com/golang/net",
		},
		{
			name:    "subdirectory module advertises the root prefix",
			target:  "http://hday.dev/pretty",
			goGet:   true,
			code:    200,
			content: "hday.dev git https://github.com/Holiday-Robot/hday-go",
		},
		{
			name:    "nested subdirectory module advertises the ancestor prefix",
			target:  "http://hday.dev/svc/kamino",
			goGet:   true,
			code:    200,
			content: "hday.dev/svc git https://github.com/Holiday-Robot/svc",
		},
		{
			name:    "hidden package adds noindex",
			target:  "http://hday.dev/secret",
			goGet:   true,
			code:    200,
			content: "hday.dev/secret git https://github.com/Holiday-Robot/secret",
			noindex: true,
		},
		{
			name:   "request without go-get is not found",
			target: "http://hday.dev/x/net",
			goGet:  false,
			code:   404,
		},
		{
			name:   "unknown package is not found",
			target: "http://hday.dev/nope",
			goGet:  true,
			code:   404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := tt.target
			if tt.goGet {
				url += "?go-get=1"
			}
			req := httptest.NewRequest("GET", url, nil)
			rec := httptest.NewRecorder()
			s.ServeHTTP(rec, req)

			if rec.Code != tt.code {
				t.Fatalf("status: got %d, want %d", rec.Code, tt.code)
			}
			if tt.code == 404 {
				return
			}

			body := rec.Body.String()
			want := `<meta name="go-import" content="` + tt.content + `"/>`
			if !strings.Contains(body, want) {
				t.Errorf("body does not contain %q\n got: %s", want, body)
			}
			// Guard against the root package regressing to a trailing slash.
			if strings.Contains(body, `content="hday.dev/ `) {
				t.Errorf("advertised prefix has a trailing slash: %s", body)
			}

			hasNoindex := strings.Contains(body, `<meta name="robots" content="noindex"/>`)
			if hasNoindex != tt.noindex {
				t.Errorf("noindex: got %v, want %v", hasNoindex, tt.noindex)
			}
		})
	}
}
