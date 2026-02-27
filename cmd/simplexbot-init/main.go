package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

type projectConfig struct {
	Module    string
	Name      string
	WSURL     string
	SDKModule string
}

type initConfig struct {
	OutDir  string
	Force   bool
	Project projectConfig
}

func main() {
	cfg, err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() (initConfig, error) {
	var (
		module    string
		name      string
		outDir    string
		wsURL     string
		sdkModule string
		force     bool
	)

	flag.StringVar(&module, "module", "", "go module path for the new bot project (required)")
	flag.StringVar(&name, "name", "", "bot display name (default: derived from module)")
	flag.StringVar(&outDir, "out", "./simplex-bot", "output directory")
	flag.StringVar(&wsURL, "ws", "ws://localhost:5225", "SimpleX websocket URL")
	flag.StringVar(&sdkModule, "sdk-module", "github.com/Malomalsky/go-simplex", "go-simplex module import path")
	flag.BoolVar(&force, "force", false, "overwrite existing files")
	flag.Parse()

	module = strings.TrimSpace(module)
	if module == "" {
		return initConfig{}, fmt.Errorf("-module is required")
	}
	if strings.ContainsAny(module, " \t\r\n") {
		return initConfig{}, fmt.Errorf("module path contains whitespace: %q", module)
	}
	if strings.TrimSpace(wsURL) == "" {
		return initConfig{}, fmt.Errorf("-ws is required")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = defaultNameFromModule(module)
	}
	if name == "" {
		return initConfig{}, fmt.Errorf("could not derive bot name from module %q", module)
	}

	outDir = strings.TrimSpace(outDir)
	if outDir == "" {
		return initConfig{}, fmt.Errorf("-out is required")
	}

	cfg := initConfig{
		OutDir: outDir,
		Force:  force,
		Project: projectConfig{
			Module:    module,
			Name:      name,
			WSURL:     wsURL,
			SDKModule: strings.TrimSpace(sdkModule),
		},
	}
	if cfg.Project.SDKModule == "" {
		return initConfig{}, fmt.Errorf("-sdk-module is required")
	}
	return cfg, nil
}

func run(cfg initConfig) error {
	files, err := buildProjectFiles(cfg.Project)
	if err != nil {
		return err
	}
	if err := writeProject(cfg.OutDir, files, cfg.Force); err != nil {
		return err
	}
	printSummary(cfg.OutDir, cfg.Project)
	return nil
}

func defaultNameFromModule(module string) string {
	parts := strings.Split(module, "/")
	last := strings.TrimSpace(parts[len(parts)-1])
	last = strings.Trim(last, ".")
	if last == "" {
		return "SimpleX Bot"
	}
	last = strings.ReplaceAll(last, "-", " ")
	last = strings.ReplaceAll(last, "_", " ")
	fields := strings.Fields(last)
	if len(fields) == 0 {
		return "SimpleX Bot"
	}
	for i := range fields {
		fields[i] = strings.ToUpper(fields[i][:1]) + fields[i][1:]
	}
	return strings.Join(fields, " ")
}

func buildProjectFiles(project projectConfig) (map[string]string, error) {
	data := map[string]any{
		"Module":    project.Module,
		"Name":      project.Name,
		"WSURL":     project.WSURL,
		"SDKModule": project.SDKModule,
	}
	files := map[string]string{
		"go.mod":     goModTemplate,
		"main.go":    mainTemplate,
		"README.md":  readmeTemplate,
		".gitignore": gitignoreTemplate,
	}

	rendered := make(map[string]string, len(files))
	for path, tpl := range files {
		content, err := renderTemplate(path, tpl, data)
		if err != nil {
			return nil, err
		}
		rendered[path] = content
	}
	return rendered, nil
}

func renderTemplate(name, src string, data map[string]any) (string, error) {
	tpl, err := template.New(name).Funcs(template.FuncMap{
		"quote": strconv.Quote,
	}).Parse(src)
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", name, err)
	}

	var b strings.Builder
	if err := tpl.Execute(&b, data); err != nil {
		return "", fmt.Errorf("execute template %s: %w", name, err)
	}
	return b.String(), nil
}

func writeProject(outDir string, files map[string]string, force bool) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, rel := range paths {
		abs := filepath.Join(outDir, rel)
		if !force {
			if _, err := os.Stat(abs); err == nil {
				return fmt.Errorf("file already exists: %s (use -force to overwrite)", abs)
			} else if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("check file %s: %w", abs, err)
			}
		}

		if err := os.WriteFile(abs, []byte(files[rel]), 0o644); err != nil {
			return fmt.Errorf("write file %s: %w", abs, err)
		}
	}
	return nil
}

func printSummary(outDir string, project projectConfig) {
	fmt.Printf("Initialized %s in %s\n\n", project.Name, outDir)
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", outDir)
	fmt.Println("  go mod tidy")
	fmt.Println("  go run .")
}

const goModTemplate = `module {{.Module}}

go 1.23
`

const mainTemplate = `package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"{{.SDKModule}}/sdk/bot"
	"{{.SDKModule}}/sdk/client"
	"{{.SDKModule}}/sdk/transport/ws"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	router := bot.NewTextRouter()
	if err := router.EnablePerContactRateLimit(20, time.Minute); err != nil {
		log.Fatalf("enable rate limit: %v", err)
	}

	if err := router.OnWithDescription("help", "show available commands", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		lines := router.HelpLines()
		if len(lines) == 0 {
			return cmd.Reply(ctx, cli, "no commands")
		}
		return cmd.Reply(ctx, cli, strings.Join(lines, "\n"))
	}); err != nil {
		log.Fatalf("register help command: %v", err)
	}

	if err := router.OnWithDescription("ping", "health check", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		return cmd.Reply(ctx, cli, "pong")
	}); err != nil {
		log.Fatalf("register ping command: %v", err)
	}

	if err := router.OnWithDescription("echo", "echo input text", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		argv, err := cmd.Argv()
		if err != nil {
			return cmd.Reply(ctx, cli, "invalid args: "+err.Error())
		}
		if len(argv) == 0 {
			return cmd.Reply(ctx, cli, "usage: /echo <text>")
		}
		return cmd.Reply(ctx, cli, strings.Join(argv, " "))
	}); err != nil {
		log.Fatalf("register echo command: %v", err)
	}

	router.OnUnknown(func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		return cmd.Reply(ctx, cli, "unknown command, try /help")
	})

	router.OnRateLimited(func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		return cmd.Reply(ctx, cli, "rate limit exceeded, try again later")
	})

	err := bot.RunWebSocketWithReconnect(
		ctx,
		{{quote .WSURL}},
		[]ws.Option{
			// For remote deployment, add: ws.WithRequireWSS(true),
			ws.WithReadLimit(16 << 20),
		},
		[]client.Option{
			client.WithStrictResponses(false),
		},
		func(cli *client.Client) (bot.Runner, error) {
			boot, err := cli.BootstrapBot(ctx)
			if err != nil {
				return nil, err
			}
			log.Printf("bot address: %s", boot.Address)

			rt, err := bot.NewRuntime(cli)
			if err != nil {
				return nil, err
			}
			rt.OnError(func(ctx context.Context, err error) {
				log.Printf("runtime error: %v", err)
			})
			bot.OnDirectCommands(rt, router)
			return rt, nil
		},
		bot.WithReconnectBackoff(1*time.Second, 20*time.Second),
		bot.WithReconnectMaxConsecutiveFailures(0),
		bot.WithReconnectErrorHandler(func(err error) {
			log.Printf("reconnect: %v", err)
		}),
	)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("bot stopped with error: %v", err)
	}
}
`

const readmeTemplate = `# {{.Name}}

Bot project scaffolded with ` + "`go-simplex`" + `.

## Run

1. Start SimpleX CLI with websocket API:

   ` + "```bash" + `
   simplex-chat -p 5225
   ` + "```" + `

2. Install deps and run bot:

   ` + "```bash" + `
   go mod tidy
   go run .
   ` + "```" + `

## Commands

- ` + "`/help`" + `
- ` + "`/ping`" + `
- ` + "`/echo <text>`" + `

## Security defaults

- per-contact command rate limit
- websocket read size limit
- reconnect supervisor with backoff
- forward-compatible response mode for upstream additions (` + "`WithStrictResponses(false)`" + `)

## Official SimpleX docs

- Bot overview: https://github.com/simplex-chat/simplex-chat/tree/stable/bots
- Bot API commands: https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/COMMANDS.md
- Bot API events: https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/EVENTS.md
- Bot API types: https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/TYPES.md
- Official TypeScript SDK: https://github.com/simplex-chat/simplex-chat/blob/stable/packages/simplex-chat-client/typescript/README.md
`

const gitignoreTemplate = `# Build
/bin/
/dist/
*.test
*.out
coverage.out

# Env
.env
.env.*
!.env.example

# IDE
.idea/
.vscode/
.DS_Store

# Agent traces
.codex/
.claude/
.aider*
.cursor/
.mcp/
*.chatlog
*.promptlog
`
