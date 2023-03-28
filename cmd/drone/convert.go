package drone

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/drone/go-convert/convert/drone"
	"github.com/drone/go-convert/convert/harness/downgrader"

	"github.com/alecthomas/kingpin/v2"
)

type convertCommand struct {
	input  string
	output string

	name       string
	proj       string
	org        string
	repoName   string
	repoConn   string
	kubeName   string
	kubeConn   string
	dockerConn string

	downgrade   bool
	beforeAfter bool
}

func (c *convertCommand) run(ctx *kingpin.ParseContext) error {
	// open the drone yaml
	before, err := ioutil.ReadFile(c.input)
	if err != nil {
		return err
	}

	// convert the pipeline yaml from the drone
	// format to the harness yaml format.
	converter := drone.New(
		drone.WithDockerhub(c.dockerConn),
		drone.WithKubernetes(c.kubeName, c.kubeConn),
	)
	after, err := converter.ConvertBytes(before)
	if err != nil {
		return err
	}

	// downgrade from the v1 harness yaml format
	// to the v0 harness yaml format.
	if c.downgrade {
		// downgrade to the v0 yaml
		d := downgrader.New(
			downgrader.WithCodebase(c.repoName, c.repoConn),
			downgrader.WithDockerhub(c.dockerConn),
			downgrader.WithKubernetes(c.kubeName, c.kubeConn),
			downgrader.WithName(c.name),
			downgrader.WithOrganization(c.org),
			downgrader.WithProject(c.proj),
		)
		after, err = d.Downgrade(after)
		if err != nil {
			return err
		}
	}

	if c.beforeAfter {
		// if the original yaml has separator and terminator
		// lines, strip these before showing the before / after
		before = bytes.TrimPrefix(before, []byte("---\n"))
		before = bytes.TrimSuffix(before, []byte("...\n"))
		before = bytes.TrimSuffix(before, []byte("..."))

		os.Stdout.WriteString("---\n")
		os.Stdout.Write(before)
		os.Stdout.WriteString("\n---\n")
	}

	if c.output == "" || c.output == "-" {
		// write the output to the console
		os.Stdout.Write(after)
		return nil
	} else {
		// write the output to the output file
		return ioutil.WriteFile(c.output, after, 0644)
	}
}

// helper function registers the convert command
func registerConvert(app *kingpin.CmdClause) {
	c := new(convertCommand)

	cmd := app.Command("convert", "converts a drone yaml").
		Action(c.run)

	cmd.Arg("input", "path to the drone yaml").
		Default(".drone.yml").
		StringVar(&c.input)

	cmd.Arg("output", "path to save the converted yaml").
		StringVar(&c.output)

	cmd.Flag("downgrade", "downgrade to the legacy yaml format").
		Default("true").
		BoolVar(&c.downgrade)

	cmd.Flag("before-after", "print the before and after").
		BoolVar(&c.beforeAfter)

	cmd.Flag("org", "harness organization").
		Default("default").
		StringVar(&c.org)

	cmd.Flag("project", "harness project").
		Default("default").
		StringVar(&c.proj)

	cmd.Flag("pipeline", "harness pipeline name").
		Default("default").
		StringVar(&c.name)

	cmd.Flag("repo-connector", "repository connector").
		Default("default").
		StringVar(&c.repoConn)

	cmd.Flag("repo-name", "repository name").
		Default("default").
		StringVar(&c.repoName)

	cmd.Flag("kube-connector", "kubernetes connector").
		Default("default").
		StringVar(&c.kubeConn)

	cmd.Flag("kube-namespace", "kubernets namespace").
		Default("default").
		StringVar(&c.kubeName)

	cmd.Flag("docker-connector", "dockerhub connector").
		Default("default").
		StringVar(&c.kubeName)
}
