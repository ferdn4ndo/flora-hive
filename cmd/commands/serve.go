package commands

import (
	"context"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"flora-hive/internal/controllers/routes"
	"flora-hive/internal/controllers/routes/middlewares"
	"flora-hive/lib"
)

// ServeCommand runs the HTTP server.
type ServeCommand struct{}

func (s *ServeCommand) Short() string { return "serve application" }

func (s *ServeCommand) Setup(cmd *cobra.Command) {}

func (s *ServeCommand) Run() lib.CommandRunner {
	return func(
		mw middlewares.Middlewares,
		env lib.Env,
		router lib.RequestHandler,
		route routes.Routes,
		logger lib.Logger,
		lc fx.Lifecycle,
	) *http.Server {
		server := &http.Server{
			Addr:              ":" + env.ServerPort,
			ReadHeaderTimeout: 10 * time.Second,
			Handler:           router.Gin,
		}

		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				lib.NewSentryHandler(logger, env)
				mw.Setup()
				route.Setup()
				go func() {
					logger.Info("Running server on port " + env.ServerPort)
					if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						logger.Error("Server error: " + err.Error())
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				if err := server.Shutdown(ctx); err != nil {
					logger.Error("Server forced to shutdown: " + err.Error())
				}
				return nil
			},
		})
		return server
	}
}

// NewServeCommand constructs ServeCommand.
func NewServeCommand() *ServeCommand { return &ServeCommand{} }
