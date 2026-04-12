package main

import (
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"flora-hive/cmd/commands"
)

var rootCmd = &cobra.Command{
	Use:              "flora-hive",
	Short:            "Flora Hive API",
	TraverseChildren: true,
}

// App is the Cobra root command wrapper.
type App struct {
	*cobra.Command
}

// NewApp wires subcommands.
func NewApp() App {
	cmd := App{Command: rootCmd}
	cmd.AddCommand(commands.Commands(CommonModules)...)
	return cmd
}

// RootApp is the default CLI entry.
var RootApp = NewApp()

func main() {
	_ = godotenv.Load()
	if err := RootApp.Execute(); err != nil {
		return
	}
}
