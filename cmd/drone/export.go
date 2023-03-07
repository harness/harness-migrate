package drone

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/harness/harness-migrate/internal/migrate/drone"
	"github.com/harness/harness-migrate/internal/migrate/drone/repo"
	"github.com/harness/harness-migrate/internal/tracer"

	"github.com/alecthomas/kingpin/v2"
	"golang.org/x/exp/slog"
)

type exportCommand struct {
	debug bool
	trace bool
	file  string

	Driver     string
	Datasource string
	namespace  string
}

func (c *exportCommand) run(*kingpin.ParseContext) error {

	// create the logger
	log := createLogger(c.debug)

	// attach the logger to the context
	ctx := context.Background()
	ctx = slog.NewContext(ctx, log)

	droneRepo, err := repo.NewRepository(c.Driver, c.Datasource)
	if err != nil {
		return err
	}
	// create the tracer
	tracer_ := tracer.New()
	defer tracer_.Close()

	// extract the data
	exporter := &drone.Exporter{
		Repository: droneRepo,
		Namespace:  c.namespace,
		Tracer:     tracer_,
	}
	data, err := exporter.Export(ctx)
	if err != nil {
		return err
	}

	//if no file path is provided, write the data export
	//to stdout.
	if c.file == "" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	// else write the data export to the file.
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(c.file, file, 0644)
}

// helper function registers the export command
func registerExport(app *kingpin.CmdClause) {
	c := new(exportCommand)

	cmd := app.Command("export", "export drone data").
		Action(c.run)

	cmd.Arg("save", "save the output to a file").
		StringVar(&c.file)

	cmd.Flag("namespace", "drone namespace").
		Required().
		Envar("DRONE_NAMESPACE").
		StringVar(&c.namespace)

	cmd.Flag("driver", "drone db type").
		Default("sqlite3").
		StringVar(&c.Driver)

	cmd.Flag("datasource", "drone database datasource").
		Envar("DRONE_DATABASE_DATASOURCE").
		Default("database.sqlite3").
		StringVar(&c.Datasource)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)

	cmd.Flag("trace", "enable trace logging").
		BoolVar(&c.trace)
}
