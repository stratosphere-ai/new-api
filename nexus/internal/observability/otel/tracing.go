package otel

import "context"

// Init installs a global tracer and meter provider using the given OTLP
// endpoint. Called once from cmd/server/main.go.
//
// TODO: use go.opentelemetry.io/otel/sdk with OTLP http exporter.
func Init(ctx context.Context, endpoint string) (shutdown func(context.Context) error, err error) {
	_ = ctx
	_ = endpoint
	return func(context.Context) error { return nil }, nil
}
