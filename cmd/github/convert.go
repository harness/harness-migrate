package github

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kingpin/v2"
	"github.com/drone/go-convert/convert/github"
	"github.com/drone/go-convert/convert/harness/downgrader"
)

const filePerm = 0644

type convertCommand struct {
	path    string
	version string
	output  string
}

func (c *convertCommand) run(ctx *kingpin.ParseContext) error {
	fileInfo, err := os.Stat(c.path)
	if err != nil {
		return err
	}

	if fileInfo.IsDir() {
		if c.output == "" {
			return fmt.Errorf("output directory is required when input is a directory")
		}
		if err := os.MkdirAll(filepath.Join(c.output, "harness"), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		files, err := os.ReadDir(c.path)
		if err != nil {
			return fmt.Errorf("failed to read directory: %w", err)
		}

		for _, file := range files {
			inputPath := filepath.Join(c.path, file.Name())
			outputPath := filepath.Join(c.output, "harness", file.Name())

			if filepath.Ext(file.Name()) != ".yaml" && filepath.Ext(file.Name()) != ".yml" {
				fmt.Printf("skipping non-YAML file %s\n", inputPath)
				continue
			}

			if err := convertFile(inputPath, outputPath, c.version); err != nil {
				return fmt.Errorf("failed to convert file %s: %w", inputPath, err)
			}

		}
	} else {
		if err := convertFile(c.path, c.output, c.version); err != nil {
			return fmt.Errorf("failed to convert file %s: %w", c.path, err)
		}
	}

	return nil
}

func convertFile(inputPath string, outputPath string, version string) error {
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		return err
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("input path is a directory: %s", inputPath)
	}

	file, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", inputPath, err)
	}

	converter := github.New()
	yaml, err := converter.ConvertBytes(file)
	if err != nil {
		return fmt.Errorf("failed to convert file %s: %w", inputPath, err)
	}

	if version == "v1" {
		d := downgrader.New()
		yaml, err = d.Downgrade(yaml)
		if err != nil {
			return fmt.Errorf("failed to downgrade file %s to v1: %w", inputPath, err)
		}
	}

	if _, err := os.Stdout.Write(yaml); err != nil {
		return fmt.Errorf("failed to write YAML to stdout: %w", err)
	}

	if outputPath != "" {
		if err := os.WriteFile(outputPath, yaml, filePerm); err != nil {
			return fmt.Errorf("failed to write YAML to file %s: %w", outputPath, err)
		}
	}

	return nil
}

// helper function registers the convert command
func registerConvert(app *kingpin.CmdClause) {
	c := new(convertCommand)

	cmd := app.Command("convert", "convert a github yaml").
		Action(c.run)

	cmd.Arg("path", "path to github yaml directory or file").
		Default(".github/workflows/main.yml").
		StringVar(&c.path)

	cmd.Flag("version", "harness yaml version, v1 or v2").
		Default("v2").
		StringVar(&c.version)

	cmd.Flag("output", "output location to write file(s) to").
		StringVar(&c.output)
}
