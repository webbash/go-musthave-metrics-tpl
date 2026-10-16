package config

import "fmt"

const (
	HTTP               = "http"
	GRPC               = "grpc"
	DefaultGRPCAddress = "localhost:8081"
)

// ValidateTransport rejects unsupported choices before allocating resources.
func ValidateTransport(transport string) error {
	if transport != HTTP && transport != GRPC {
		return fmt.Errorf("unknown transport %q: expected http or grpc", transport)
	}
	return nil
}
