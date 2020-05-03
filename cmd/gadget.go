package main

import (
	"context"
	"flag"
	"github.com/maruel/subcommands"
	"github.com/srabraham/gerritgadget/internal/cqdepend"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/api/gerrit"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/hardcoded/chromeinfra"
	"log"
	"os"
)

type getCqDependRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	cls       string
}

func (c *getCqDependRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	flag.Parse()

	ctx := context.Background()
	authOpts, err := c.authFlags.Options()
	if err != nil {
		log.Printf("authFlags.Options: %v", err)
		return 1
	}
	authenticator := auth.NewAuthenticator(ctx, auth.SilentLogin, authOpts)

	client, err := authenticator.Client()
	if err != nil {
		log.Printf("authenticator.Client: %v", err)
		return 1
	}

	if err = cqdepend.UpdateCqDepend(client, c.cls); err != nil {
		log.Printf("UpdateCqDepend: %v", err)
		return 1
	}
	return 0
}

func cqDepend(authOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "cq-depend --cls=chromium:123,chrome-internal:234",
		ShortDesc: "Does some stuff",
		LongDesc:  "Does some stuff",
		CommandRun: func() subcommands.CommandRun {
			c := &getCqDependRun{}
			c.authFlags = authcli.Flags{}
			c.authFlags.Register(c.GetFlags(), authOpts)
			c.Flags.StringVar(&c.cls, "cls", "", "")
			return c
		},
	}
}

func GetApplication(authOpts auth.Options) *cli.Application {
	return &cli.Application{
		Name: "test_planner",
		Context: func(ctx context.Context) context.Context {
			return ctx
		},
		Commands: []*subcommands.Command{
			authcli.SubcommandInfo(authOpts, "auth-info", false),
			authcli.SubcommandLogin(authOpts, "auth-login", false),
			authcli.SubcommandLogout(authOpts, "auth-logout", false),
			cqDepend(authOpts),
		},
	}
}

func main() {
	opts := chromeinfra.DefaultAuthOptions()
	opts.Scopes = []string{gerrit.OAuthScope, auth.OAuthScopeEmail}
	app := GetApplication(opts)
	os.Exit(subcommands.Run(app, nil))
}
