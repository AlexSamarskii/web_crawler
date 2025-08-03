// cmd/root.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "crawler-cli",
	Short: "Многопоточный веб-краулер на Go",
	Long: `Crawler — это инструмент для обхода сайтов, 
извлечения ссылок и сохранения данных в MongoDB.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Usage()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
