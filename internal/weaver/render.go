package weaver

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
)

var templateDir = os.Getenv("WEAVER_TEMPLATE_DIR")

func init() {
	if templateDir == "" {
		templateDir = "/opt/weaver/templates"
	}
}

func RenderDashboards(ctx context.Context, entity string, schemas []string) ([]string, error) {
	schemaDir, err := os.MkdirTemp("/tmp", "schema-")
	if err != nil {
		return nil, fmt.Errorf("creating temporary directory: %w", err)
	}

	for idx, contents := range schemas {
		if len(contents) == 0 {
			continue
		}
		err := os.WriteFile(path.Join(schemaDir, fmt.Sprintf("schema-%d.yaml", idx)), []byte(contents), 0644)
		if err != nil {
			return nil, fmt.Errorf("writing schema file: %w", err)
		}
	}

	manifestFile, err := os.Create(path.Join(schemaDir, "registry_manifest.yaml"))
	if err != nil {
		return nil, fmt.Errorf("writing manifest file: %w", err)
	}
	defer manifestFile.Close()
	fmt.Fprintf(manifestFile, `name: %s
description: Custom Conventions
semconv_version: 0.0.1
schema_base_url: github.com/ArthurSens/demo-weaver-for-dashboarding/semconv
`, entity)

	outdir := path.Join(schemaDir, "output")
	exe := exec.Command("weaver")
	exe.Args = []string{
		"weaver", "registry", "generate", "-r", schemaDir, "-t", templateDir, "dashboards", outdir,
	}
	exe.Stdout = os.Stdout
	exe.Stderr = os.Stderr
	slog.DebugContext(ctx, "running weaver", "args", exe.Args)
	if err := exe.Run(); err != nil {
		return nil, fmt.Errorf("generating dashboards using weaver: %w", err)
	}
	files, err := os.ReadDir(outdir)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}
	slog.DebugContext(ctx, "generated dashboards", "total", len(files))
	dashboards := make([]string, len(files))
	for idx, f := range files {
		contents, err := os.ReadFile(path.Join(outdir, f.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading dashboard: %w", err)
		}
		dashboards[idx] = string(contents)

	}
	return dashboards, nil
}
