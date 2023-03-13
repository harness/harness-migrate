package drone

import (
	"os"

	"github.com/drone/go-convert/convert/drone"

	"github.com/alecthomas/kingpin/v2"
)

type convertCommand struct {
	path string
}

func (c *convertCommand) run(*kingpin.ParseContext) error {
	file, err := os.ReadFile(c.path)
	if err != nil {
		return err
	}
	converter := drone.New()
	yaml, err := converter.ConvertBytes(file)
	if err != nil {
		return err
	}
	os.Stdout.Write(yaml)
	return nil
}

// helper function registers the convert command
func registerConvert(app *kingpin.CmdClause) {
	c := new(convertCommand)

	cmd := app.Command("convert", "convert a drone yaml").
		Action(c.run)

	cmd.Arg("path", "path to drone yaml").
		Default(".drone.yml").
		StringVar(&c.path)
}
