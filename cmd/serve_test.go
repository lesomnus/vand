package cmd

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lesomnus/vand/cmd/config"
	"github.com/lesomnus/vand/internal/vand"
)

func TestServeMux(t *testing.T) {
	mux := newServeMux(vand.NewServer(config.PackageConfig{
		"": {Target: "https://github.com/Holiday-Robot/hday-go"},
	}))

	t.Run("plain root is an unlogged health check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://hday.dev/", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != 200 {
			t.Fatalf("status: got %d, want 200", rec.Code)
		}
		if body := rec.Body.String(); body != "" {
			t.Errorf("health check should have an empty body, got %q", body)
		}
	})

	t.Run("root go-get is delegated to the vanity handler", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://hday.dev/?go-get=1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != 200 {
			t.Fatalf("status: got %d, want 200", rec.Code)
		}
		want := `<meta name="go-import" content="hday.dev git https://github.com/Holiday-Robot/hday-go"/>`
		if body := rec.Body.String(); !strings.Contains(body, want) {
			t.Errorf("body does not contain %q\n got: %s", want, body)
		}
	})
}
