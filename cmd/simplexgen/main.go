package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Malomalsky/go-simplex/internal/spec"
)

func main() {
	var (
		inputPath  string
		outputPath string
		pkgName    string
	)

	flag.StringVar(&inputPath, "input", "spec/upstream/COMMANDS.md", "path to COMMANDS.md")
	flag.StringVar(&outputPath, "output", "sdk/command/generated_catalog.go", "generated go file path")
	flag.StringVar(&pkgName, "package", "command", "package name for generated file")
	flag.Parse()

	if err := run(inputPath, outputPath, pkgName); err != nil {
		fmt.Fprintf(os.Stderr, "simplexgen: %v\n", err)
		os.Exit(1)
	}
}

func run(inputPath, outputPath, pkgName string) error {
	in, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer in.Close()

	doc, err := spec.ParseCommandsMarkdown(in)
	if err != nil {
		return fmt.Errorf("parse commands markdown: %w", err)
	}

	code, err := spec.RenderCatalogGo(doc, pkgName)
	if err != nil {
		return fmt.Errorf("render go catalog: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(outputPath, code, 0o644); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}
	return nil
}
