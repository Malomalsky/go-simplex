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

const (
	projectTemplateBasic      = "basic"
	projectTemplateModeration = "moderation"
)

type projectConfig struct {
	Module    string
	Name      string
	WSURL     string
	SDKModule string
	Template  string
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
		template  string
		force     bool
	)

	flag.StringVar(&module, "module", "", "go module path for the new bot project (required)")
	flag.StringVar(&name, "name", "", "bot display name (default: derived from module)")
	flag.StringVar(&outDir, "out", "./simplex-bot", "output directory")
	flag.StringVar(&wsURL, "ws", "ws://localhost:5225", "SimpleX websocket URL")
	flag.StringVar(&sdkModule, "sdk-module", "github.com/Malomalsky/go-simplex", "go-simplex module import path")
	flag.StringVar(&template, "template", projectTemplateBasic, "project template: basic|moderation")
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
			Template:  strings.TrimSpace(template),
		},
	}
	if cfg.Project.SDKModule == "" {
		return initConfig{}, fmt.Errorf("-sdk-module is required")
	}
	if cfg.Project.Template == "" {
		cfg.Project.Template = projectTemplateBasic
	}
	if cfg.Project.Template != projectTemplateBasic && cfg.Project.Template != projectTemplateModeration {
		return initConfig{}, fmt.Errorf("unsupported template: %q (expected: %s, %s)", cfg.Project.Template, projectTemplateBasic, projectTemplateModeration)
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
	files, err := filesForTemplate(project.Template)
	if err != nil {
		return nil, err
	}

	data := map[string]any{
		"Module":    project.Module,
		"Name":      project.Name,
		"WSURL":     project.WSURL,
		"SDKModule": project.SDKModule,
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

func filesForTemplate(template string) (map[string]string, error) {
	base := map[string]string{
		"go.mod":     goModTemplate,
		".gitignore": gitignoreTemplate,
	}

	switch template {
	case projectTemplateBasic:
		base["main.go"] = mainTemplateBasic
		base["README.md"] = readmeTemplateBasic
		return base, nil
	case projectTemplateModeration:
		base["main.go"] = mainTemplateModeration
		base["README.md"] = readmeTemplateModeration
		return base, nil
	default:
		return nil, fmt.Errorf("unsupported template: %q", template)
	}
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
	fmt.Printf("Template: %s\n\n", project.Template)
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", outDir)
	fmt.Println("  go mod tidy")
	fmt.Println("  go run .")
}

const goModTemplate = `module {{.Module}}

go 1.23
`

const mainTemplateBasic = `package main

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

const mainTemplateModeration = `package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"{{.SDKModule}}/sdk/bot"
	"{{.SDKModule}}/sdk/client"
	"{{.SDKModule}}/sdk/transport/ws"
)

type denyList struct {
	mu    sync.RWMutex
	words map[string]struct{}
}

func newDenyList(initial []string) *denyList {
	d := &denyList{words: make(map[string]struct{}, len(initial))}
	for _, w := range initial {
		d.Add(w)
	}
	return d
}

func (d *denyList) Add(word string) {
	word = normalizeWord(word)
	if word == "" {
		return
	}
	d.mu.Lock()
	d.words[word] = struct{}{}
	d.mu.Unlock()
}

func (d *denyList) Remove(word string) bool {
	word = normalizeWord(word)
	if word == "" {
		return false
	}
	d.mu.Lock()
	_, exists := d.words[word]
	if exists {
		delete(d.words, word)
	}
	d.mu.Unlock()
	return exists
}

func (d *denyList) List() []string {
	d.mu.RLock()
	out := make([]string, 0, len(d.words))
	for w := range d.words {
		out = append(out, w)
	}
	d.mu.RUnlock()
	sort.Strings(out)
	return out
}

func (d *denyList) Match(text string) (string, bool) {
	normalized := strings.ToLower(text)
	d.mu.RLock()
	defer d.mu.RUnlock()
	for w := range d.words {
		if strings.Contains(normalized, w) {
			return w, true
		}
	}
	return "", false
}

func normalizeWord(word string) string {
	return strings.ToLower(strings.TrimSpace(word))
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	list := newDenyList([]string{"spam", "scam"})

	router := bot.NewTextRouter()
	if err := router.EnablePerContactRateLimit(20, time.Minute); err != nil {
		log.Fatalf("enable rate limit: %v", err)
	}

	if err := router.OnWithDescription("help", "show available commands", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		return cmd.Reply(ctx, cli, strings.Join(router.HelpLines(), "\n"))
	}); err != nil {
		log.Fatalf("register help command: %v", err)
	}

	if err := router.OnWithDescription("addword", "add word to deny-list", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		word, ok := cmd.Arg(0)
		if !ok {
			return cmd.Reply(ctx, cli, "usage: /addword <word>")
		}
		list.Add(word)
		return cmd.Reply(ctx, cli, "added: "+normalizeWord(word))
	}); err != nil {
		log.Fatalf("register addword command: %v", err)
	}

	if err := router.OnWithDescription("delword", "remove word from deny-list", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		word, ok := cmd.Arg(0)
		if !ok {
			return cmd.Reply(ctx, cli, "usage: /delword <word>")
		}
		if !list.Remove(word) {
			return cmd.Reply(ctx, cli, "not found: "+normalizeWord(word))
		}
		return cmd.Reply(ctx, cli, "removed: "+normalizeWord(word))
	}); err != nil {
		log.Fatalf("register delword command: %v", err)
	}

	if err := router.OnWithDescription("words", "show deny-list", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		words := list.List()
		if len(words) == 0 {
			return cmd.Reply(ctx, cli, "deny-list is empty")
		}
		return cmd.Reply(ctx, cli, "deny-list: "+strings.Join(words, ", "))
	}); err != nil {
		log.Fatalf("register words command: %v", err)
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
			bot.OnDirectText(rt, func(ctx context.Context, cli *client.Client, msg bot.DirectTextMessage) error {
				if strings.HasPrefix(strings.TrimSpace(msg.Text), "/") {
					return nil
				}
				if word, matched := list.Match(msg.Text); matched {
					return msg.Reply(ctx, cli, "message blocked by moderation rule: "+word)
				}
				return nil
			})
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

const readmeTemplateBasic = `# {{.Name}}

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

const readmeTemplateModeration = `# {{.Name}}

Bot project scaffolded with ` + "`go-simplex`" + `, moderation template.

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
- ` + "`/addword <word>`" + `
- ` + "`/delword <word>`" + `
- ` + "`/words`" + `

Default deny-list contains: ` + "`spam`" + ` and ` + "`scam`" + `.

## Behavior

- commands manage in-memory deny-list
- non-command direct messages are checked against deny-list
- if blocked word is found, bot replies with moderation warning

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
