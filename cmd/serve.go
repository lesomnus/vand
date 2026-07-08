package cmd

import (
	"context"
	"net/http"

	"github.com/lesomnus/otx"
	"github.com/lesomnus/otx/otxhttp"
	"github.com/lesomnus/vand/internal/vand"
	"github.com/lesomnus/xli"
)

func NewCmdServe() *xli.Command {
	return &xli.Command{
		Name:  "serve",
		Brief: "start the vanity import server",

		Handler: xli.OnRun(func(ctx context.Context, cmd *xli.Command, next xli.Next) error {
			c := use_config.Must(ctx)
			s := vand.NewServer(c.Packages)

			x := otx.From(ctx)
			h := otxhttp.NewHandler(x, otxhttp.BoundaryLogger()(s), "vand")

			return http.ListenAndServe(c.Server.Addr, newServeMux(h))
		}),
	}
}

// newServeMux wires the health check and the vanity handler. The root path is
// both the health check and the vanity endpoint for the root package, so it is
// delegated to the vanity handler only when it is a go-get request; a plain
// GET / stays a lightweight, unlogged 200.
func newServeMux(vanity http.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("go-get") == "1" {
			vanity.ServeHTTP(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/", vanity)
	return mux
}
