package commands

import (
	"github.com/spf13/cobra"

	"flora-hive/internal/bootstrapauth"
	"flora-hive/lib"
)

// BootstrapAuthCommand wires bootstrap:auth (no Fx / DB), matching userver-filemgr.
func BootstrapAuthCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "bootstrap:auth",
		Short: "Create uServer-Auth system and optional first admin (HTTP only)",
		Long: `Calls uServer-Auth: POST /auth/system (Token: USERVER_AUTH_SYSTEM_CREATION_TOKEN or SYSTEM_CREATION_TOKEN)
and optionally POST /auth/register. Controlled by env; see .env.example.

Persists tokens to ENV_FILE / HIVE_ENV_FILE (default .env) when keys are empty or placeholders there,
unless HIVE_SKIP_PERSIST_BOOTSTRAP_ENV=1 (or FILEMGR_SKIP_PERSIST_BOOTSTRAP_ENV=1).

Skip with SKIP_USERVER_AUTH_SETUP=1 or SKIP_AUTH_BOOTSTRAP=1. Safe when variables are unset (no-op).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return bootstrapauth.Run(cmd.OutOrStdout(), lib.NewEnv())
		},
	}
}
