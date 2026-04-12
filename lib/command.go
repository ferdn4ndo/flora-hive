package lib

import "github.com/spf13/cobra"

// CommandRunner is the return type of Command.Run (fx injects dependencies).
type CommandRunner interface{}

// Command implements cobra subcommands.
type Command interface {
	Short() string
	Setup(cmd *cobra.Command)
	Run() CommandRunner
}
