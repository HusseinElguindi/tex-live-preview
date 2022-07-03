package cmd

import (
	"errors"
	"os"

	"github.com/husseinelguindi/tex-live-preview/preview"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tex-live-preview file",
	Short: "monitors for TeX file changes in the parent directory of the passed file, then recompiles (using pdflatex)",
	Args: func(_ *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("expected one argument")
		}

		f, err := os.Stat(args[0])
		if err == nil || os.IsExist(err) {
			if f.IsDir() {
				return errors.New("provided filepath must correspond to a file, not a directory")
			}
			return nil // file exists
		}
		if os.IsNotExist(err) {
			return errors.New("provided filepath does not correspond to an existing file")
		}
		return err

	},
	RunE: preview.Start,
}

func Execute() error {
	return rootCmd.Execute()
}
