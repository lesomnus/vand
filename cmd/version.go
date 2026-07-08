package cmd

import (
	"context"

	"github.com/lesomnus/vand/cmd/version"
	"github.com/lesomnus/xli"
)

func NewCmdVersion() *xli.Command {
	const Template = `VAND_VERSION=%s
VAND_GIT_REV=%s
VAND_GIT_DIRTY=%v
`
	return &xli.Command{
		Name:  "version",
		Brief: "print version information",
		Handler: xli.OnRun(func(ctx context.Context, cmd *xli.Command, next xli.Next) error {
			v := version.Get()
			cmd.Printf(Template,
				v.Version,
				v.GitRev,
				v.GitDirty,
			)
			return nil
		}),
	}
}
