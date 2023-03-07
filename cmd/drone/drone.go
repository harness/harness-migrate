package drone

import "github.com/alecthomas/kingpin/v2"

func Register(app *kingpin.Application) {
	cmd := app.Command("drone", "migrate drone data")
	//registerMigrate(cmd)
	registerExport(cmd)
	registerImport(cmd)
	registerConvert(cmd)
}
