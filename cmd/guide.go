package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	yaml "gopkg.in/yaml.v2"

	"github.com/alecthomas/chroma/quick"
	"github.com/alecthomas/template"
	"github.com/s2gatev/lingo/checker"
	"github.com/spf13/cobra"
)

func init() {
	Guide.PersistentFlags().StringVar(
		&configFile, "config", defaultConfigFilename, "config file")

	Root.AddCommand(Guide)
}

// Guide is a command handler that displays a guidebook of rules applicable
// for the current project.
var Guide = &cobra.Command{
	Use:   "guide",
	Short: "Read a guide with the lingo of the project",
	Run: func(cmd *cobra.Command, args []string) {
		configData, err := ioutil.ReadFile(configFile)
		if err != nil {
			// TODO: handle error gracefully
			panic(err)
		}

		var config Config
		if err := yaml.Unmarshal(configData, &config); err != nil {
			// TODO: handle error gracefully
			panic(err)
		}

		var checkers []checker.NodeChecker
		for slug, config := range config.Checkers {
			c := checker.Get(slug, config)
			if c == nil {
				// TODO: handle error gracefully
				panic("unknown checker: " + slug)
			}

			checkers = append(checkers, c)
		}

		configPath, err := filepath.Abs(configFile)
		if err != nil {
			// TODO: handle error gracefully
			panic(err)
		}

		project := filepath.Base(filepath.Dir(configPath))

		var items []guideItem
		for _, checker := range checkers {
			item := guideItem{
				Title:       checker.Title(),
				Description: checker.Description(),
			}

			for _, example := range checker.Examples() {
				var good bytes.Buffer
				err = quick.Highlight(&good, example.Good, "go", "html", "github")
				if err != nil {
					// TODO: handle error gracefully
					panic(err)
				}

				var bad bytes.Buffer
				err = quick.Highlight(&bad, example.Bad, "go", "html", "github")
				if err != nil {
					// TODO: handle error gracefully
					panic(err)
				}

				item.Examples = append(item.Examples, guideItemExample{
					Good: good.String(),
					Bad:  bad.String(),
				})
			}

			items = append(items, item)
		}

		dir, err := ioutil.TempDir("", "lingo")
		if err != nil {
			// TODO: handle error gracefully
			panic(err)
		}

		guide, err := os.Create(filepath.Join(dir, "guide.html"))
		if err != nil {
			// TODO: handle error gracefully
			panic(err)
		}
		defer guide.Close()

		code := `package main

import "fmt"

func main() {
	fmt.Println("Hello World!")
}
`

		var highlighted bytes.Buffer
		err = quick.Highlight(&highlighted, code, "go", "html", "github")
		if err != nil {
			// TODO: handle error gracefully
			panic(err)
		}

		data := map[string]interface{}{
			"Project": project,
			"Items":   items,
			"Code":    highlighted.String(),
		}

		if err := guideTemplate.Execute(guide, data); err != nil {
			return
		}

		if err := openBrowser("file://" + guide.Name()); err != nil {
			return
		}
	},
}

// openBrowser tries to open the URL in a browser.
func openBrowser(url string) error {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}

	return exec.Command(args[0], append(args[1:], url)...).Run()
}

type guideItemExample struct {

	// Good is an example of sticking to the rule.
	Good string

	// Bad is a counter-example that shows how to not apply the rule.
	Bad string
}

type guideItem struct {

	// Title is the title of the item.
	Title string

	// Description is the detailed description of the item.
	Description string

	// Examples is a set of examples of applying item.
	Examples []guideItemExample
}

var guideTemplate = template.Must(template.New("html").Parse(guideContent))

const guideContent = `
<!DOCTYPE html>
<html>
	<head>
		<title>{{.Project}}'s lingo</title>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
		<style>
			body {
				font-family: -apple-system, BlinkMacSystemFont,
					"Segoe UI", Helvetica, Arial, sans-serif,
					"Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol";
			}

			h1, h2 {
				padding-bottom: 10px;
				border-bottom: 1px solid #eaecef;
			}

			.page {
				margin: 0 auto;
				width: 980px;
				padding: 45px;
				border: 1px solid #ddd;
			}

			.items {
				margin-top: 50px;
			}

			.item:not(:last-child) {
				padding-bottom: 30px;
			}

			.code {
				padding: 0 10px;
				border: 1px solid #eaecef;
			}
		</style>
	</head>
	<body>
		<div class="page">
			<h1>{{.Project}}'s lingo</h1>
			<div class="items">
				{{range .Items}}
				<div class="item">
					<h2>{{.Title}}</h2>
					<p>{{.Description}}</p>

					{{range .Examples}}
					<h4>Bad</h4>
					<div class="code">{{.Bad}}</div>

					<h4>Good</h4>
					<div class="code">{{.Good}}</div>
					{{end}}
				</div>
				{{end}}
			</div>
		</div>
	</body>
</html>
`
