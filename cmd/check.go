package cmd

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"

	"github.com/s2gatev/lingo/checker"
	"github.com/s2gatev/lingo/cli"
	"github.com/s2gatev/lingo/file"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func init() {
	Check.PersistentFlags().StringVar(
		&configFile, "config", defaultConfigFilename, "config file")

	Root.AddCommand(Check)
}

// Check is a command handler that checks the lingo in a directory
// for violations.
var Check = &cobra.Command{
	Use:   "check",
	Short: "Check the lingo of all files in a directory",
	Run: func(cmd *cobra.Command, args []string) {
		configData, err := ioutil.ReadFile(configFile)
		if err != nil {
			cli.ExitError("failed to read config file: %s", configFile)
		}

		var config Config
		if err := yaml.Unmarshal(configData, &config); err != nil {
			cli.ExitError("failed to parse config file: %s", configFile)
		}

		var matchers []file.Matcher
		for _, matcher := range config.Matchers {
			matchers = append(matchers, file.Get(matcher.Type, matcher.Config))
		}
		feeder := file.NewFeeder(matchers...)

		fc := checker.NewFileChecker()
		for slug, config := range config.Checkers {
			c := checker.Get(slug, config)
			if c == nil {
				cli.ExitError("unknown checker: %s", slug)
			}

			fc.Register(c)
		}

		files, err := feeder.Feed(args[0])
		if err != nil {
			cli.ExitError("failed to process files: %s", args[0])
		}

		reports := map[string]*checker.Report{}
		fileSets := map[string]*token.FileSet{}

		for path := range files {
			reports[path] = &checker.Report{}
			fileSets[path] = token.NewFileSet()

			content, err := ioutil.ReadFile(path)
			if err != nil {
				panic(err)
			}

			file, err := parser.ParseFile(
				fileSets[path],
				path,
				nil,
				parser.ParseComments)
			if err != nil {
				cli.ExitError("failed to parse file: %s", path)
			}

			fc.Check(file, string(content), reports[path])
		}

		totalErrors := 0
		for path, report := range reports {
			if len(report.Errors) == 0 {
				continue
			}

			fmt.Println(path)
			for _, err := range report.Errors {
				position := fileSets[path].Position(err.Pos)
				fmt.Printf("\t- line %d: %s\n", position.Line, err.Message)
			}
			fmt.Println()

			totalErrors += len(report.Errors)
		}

		if totalErrors > 0 {
			cli.ExitError("%d violations found in %d files",
				totalErrors, len(reports))
		} else {
			cli.ExitOK("%d violations found in %d files",
				totalErrors, len(reports))
		}
	},
}
