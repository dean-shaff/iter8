package basecli

import (
	"github.com/spf13/cobra"
)

type GenOptionsType struct {
	// Values are user specified values used during gen
	Values []string
}

var GenOptions = GenOptionsType{}

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Render templates with values",
	Long: `
	Render templates with values`,
}

func init() {
	RootCmd.AddCommand(genCmd)
	genCmd.PersistentFlags().StringSliceVarP(&GenOptions.Values, "set", "s", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
}
