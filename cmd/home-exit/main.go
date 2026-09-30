package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/lovrozagar/home-exit/internal/config"
	"github.com/lovrozagar/home-exit/internal/proxy"
	"github.com/lovrozagar/home-exit/internal/users"
)

func main() {
	log.SetFlags(log.LstdFlags)
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "up":
		os.Exit(cmdUp(os.Args[2:]))
	case "status":
		os.Exit(cmdStatus(os.Args[2:]))
	case "users":
		os.Exit(cmdUsers(os.Args[2:]))
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `home-exit — authenticated SOCKS5 + HTTP CONNECT egress

Usage:
  home-exit up                 Start the proxy (foreground)
  home-exit status             Print config summary
  home-exit users add <name>   Create a client; prints password once
  home-exit users list         List usernames (no secrets)

Flags (all commands):
  -config path   Config file (also: HOME_EXIT_CONFIG)

Config search order:
  1. -config / HOME_EXIT_CONFIG
  2. ./home-exit.yaml
  3. ~/.config/home-exit/config.yaml

`)
}

func parseConfigFlag(args []string) (remaining []string, cfgPath string, err error) {
	fs := flag.NewFlagSet("home-exit", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	cfg := fs.String("config", "", "path to home-exit.yaml")
	if err := fs.Parse(args); err != nil {
		return nil, "", err
	}
	path, err := config.FindPath(*cfg)
	if err != nil {
		return nil, "", err
	}
	return fs.Args(), path, nil
}

func loadStore(cfgPath string) (config.Config, *users.Store, string, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return cfg, nil, "", err
	}
	usersPath := cfg.UsersFilePath(cfgPath)
	store, err := users.Open(usersPath)
	if err != nil {
		return cfg, nil, usersPath, err
	}
	return cfg, store, usersPath, nil
}

func cmdUp(args []string) int {
	_, cfgPath, err := parseConfigFlag(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	cfg, store, usersPath, err := loadStore(cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if store.Count() == 0 {
		fmt.Fprintf(os.Stderr, "no users configured — run: home-exit users add <name>\n(users file: %s)\n", usersPath)
		return 1
	}

	fmt.Print(cfg.Summary(cfgPath, usersPath))
	fmt.Println()
	fmt.Println("Connect examples (replace USER:PASS):")
	fmt.Printf("  curl -x socks5h://USER:PASS@%s https://ifconfig.me\n", cfg.Listen.SOCKS)
	fmt.Printf("  curl -x http://USER:PASS@%s https://ifconfig.me\n", cfg.Listen.HTTP)
	fmt.Println()
	fmt.Println("Auth is required. Anonymous connections are rejected.")
	fmt.Println("Ctrl+C to stop.")
	fmt.Println()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &proxy.Server{
		SOCKSAddr: cfg.Listen.SOCKS,
		HTTPAddr:  cfg.Listen.HTTP,
		Auth:      store,
		Logger:    log.Default(),
	}
	if err := srv.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func cmdStatus(args []string) int {
	_, cfgPath, err := parseConfigFlag(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	cfg, store, usersPath, err := loadStore(cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Print(cfg.Summary(cfgPath, usersPath))
	fmt.Printf("users:   %d configured\n", store.Count())
	fmt.Println("running: N/A (v1 is a foreground process; use `home-exit up`)")
	return 0
}

func cmdUsers(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: home-exit users add <name> | home-exit users list")
		return 2
	}
	switch args[0] {
	case "add":
		return cmdUsersAdd(args[1:])
	case "list":
		return cmdUsersList(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown users subcommand: %s\n", args[0])
		return 2
	}
}

func cmdUsersAdd(args []string) int {
	cfgFlag, name, err := parseUsersAddArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "usage: home-exit users add <name> [-config path]")
		return 2
	}
	cfgPath, err := config.FindPath(cfgFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	_, store, usersPath, err := loadStore(cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	plain, err := store.Add(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("created user %q\n", name)
	fmt.Printf("password (shown once): %s\n", plain)
	fmt.Printf("stored in: %s\n", usersPath)
	fmt.Println("Save the password now — only a bcrypt hash is kept on disk.")
	return 0
}


func parseUsersAddArgs(args []string) (cfgPath, name string, err error) {
	var positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-config" || a == "--config":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("missing value for %s", a)
			}
			i++
			cfgPath = args[i]
		case strings.HasPrefix(a, "-config="):
			cfgPath = strings.TrimPrefix(a, "-config=")
		case strings.HasPrefix(a, "--config="):
			cfgPath = strings.TrimPrefix(a, "--config=")
		case strings.HasPrefix(a, "-"):
			return "", "", fmt.Errorf("unknown flag: %s", a)
		default:
			positional = append(positional, a)
		}
	}
	if len(positional) != 1 || strings.TrimSpace(positional[0]) == "" {
		return "", "", fmt.Errorf("expected exactly one username")
	}
	return cfgPath, positional[0], nil
}

func cmdUsersList(args []string) int {
	_, cfgPath, err := parseConfigFlag(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	_, store, usersPath, err := loadStore(cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	names := store.List()
	if len(names) == 0 {
		fmt.Printf("(no users in %s)\n", usersPath)
		return 0
	}
	for _, n := range names {
		fmt.Println(n)
	}
	return 0
}
