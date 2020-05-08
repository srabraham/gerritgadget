package main

import (
	"context"
	"flag"
	"fmt"
	aggerrit "github.com/andygrunwald/go-gerrit"
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
		UsageLine: "apply-cq-depend --cls=chromium:123,chrome-internal:234",
		ShortDesc: "applies a fully connected graph of Cq-Depend footers for the provided CLs",
		LongDesc:  "Applies a fully connected graph of Cq-Depend footers for the provided CLs",
		CommandRun: func() subcommands.CommandRun {
			c := &getCqDependRun{}
			c.authFlags = authcli.Flags{}
			c.authFlags.Register(c.GetFlags(), authOpts)
			c.Flags.StringVar(&c.cls, "cls", "", "List of CLs to Cq-Depend, e.g. chromium:123,chrome-internal:234 (generally, short-host:change-num)")
			return c
		},
	}
}

type getCreateBranchRun struct {
	subcommands.CommandRunBase
	authFlags  authcli.Flags
	host       string
	project    string
	sourceRef  string
	destBranch string
}

func (c *getCreateBranchRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
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

	agClient, err := aggerrit.NewClient(fmt.Sprintf("https://%v-review.googlesource.com", c.host), client)
	if err != nil {
		log.Printf("failed to create Gerrit client: %v", err)
		return 1
	}

	bi, _, err := agClient.Projects.CreateBranch(c.project, c.destBranch, &aggerrit.BranchInput{Ref: c.sourceRef})
	if err != nil {
		log.Printf("failed to create branch: %v", err)
	}
	log.Printf("got branchinfo: %v", bi)
	return 0
}

func createBranch(authOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "create-branch --host=chromium --project=chromiumos/chromite --source-ref=ad7432b1897412 --dest-branch=test-branch-name",
		ShortDesc: "",
		LongDesc:  "",
		CommandRun: func() subcommands.CommandRun {
			c := &getCreateBranchRun{}
			c.authFlags = authcli.Flags{}
			c.authFlags.Register(c.GetFlags(), authOpts)
			c.Flags.StringVar(&c.host, "host", "", "")
			c.Flags.StringVar(&c.project, "project", "", "")
			c.Flags.StringVar(&c.sourceRef, "source-ref", "", "")
			c.Flags.StringVar(&c.destBranch, "dest-branch", "", "")
			return c
		},
	}
}

func GetApplication(authOpts auth.Options) *cli.Application {
	return &cli.Application{
		Name: "gerritgadgetcli",
		Context: func(ctx context.Context) context.Context {
			return ctx
		},
		Commands: []*subcommands.Command{
			authcli.SubcommandInfo(authOpts, "auth-info", false),
			authcli.SubcommandLogin(authOpts, "auth-login", false),
			authcli.SubcommandLogout(authOpts, "auth-logout", false),
			cqDepend(authOpts),
			createBranch(authOpts),
		},
	}
}

func main() {
	opts := chromeinfra.DefaultAuthOptions()
	opts.Scopes = []string{gerrit.OAuthScope, auth.OAuthScopeEmail}
	app := GetApplication(opts)
	os.Exit(subcommands.Run(app, nil))
}
