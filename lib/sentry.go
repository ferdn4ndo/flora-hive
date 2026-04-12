package lib

import "github.com/getsentry/sentry-go"

// NewSentryHandler initializes Sentry when not running locally.
func NewSentryHandler(logger Logger, env Env) {
	if env.IsLocal() {
		return
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              env.SentryDsn,
		Environment:      env.Environment,
		AttachStacktrace: true,
	})
	if err != nil {
		logger.Errorf("Sentry initialization failed: %v", err)
	}
}
