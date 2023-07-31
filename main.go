package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/teowa/azure-rest-api-variants/variant"
	"github.com/urfave/cli/v2"
)

var (
	flagOutput string
)

func main() {
	app := &cli.App{
		Name:      "azure-rest-api-variants",
		Version:   getVersion(),
		Usage:     "Variants of azure-rest-api-specs",
		UsageText: "azure-rest-api-index <command> [option]",
		Commands: []*cli.Command{
			{
				Name:      "build",
				Usage:     `Building the variant index`,
				UsageText: "azure-rest-api-variants build [option] <specdir>",
				Before: func(ctx *cli.Context) error {
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "output",
						Aliases:     []string{"o"},
						Usage:       `Output file`,
						Destination: &flagOutput,
					},
				},
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						return fmt.Errorf("the swagger spec dir not specified")
					}
					if c.NArg() > 1 {
						return fmt.Errorf("more than one arguments specified")
					}

					specDir := c.Args().First()
					if strings.HasSuffix(specDir, "specification/") {
						specDir = strings.TrimSuffix(specDir, "/")
					}
					if !strings.HasSuffix(specDir, "specification") {
						return fmt.Errorf("the swagger spec dir must be the specification folder, e.g., /home/test/go/src/github.com/azure/azure-rest-api-specs/specification")
					}
					index, err := variant.Build(specDir)
					if err != nil {
						return err
					}

					b, err := json.MarshalIndent(index, "", "\t")
					if err != nil {
						log.Fatal(err)
					}

					if flagOutput == "" {
						fmt.Println(string(b))
						return nil
					}
					return os.WriteFile(flagOutput, b, 0644)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
