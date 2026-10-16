package config

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransportConfiguration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"transport":"grpc","grpc_address":"json:8081"}`), 0600))
	for _, tc := range []struct {
		name                     string
		args                     []string
		envTransport, envAddress *string
		transport, address       string
	}{
		{name: "default", transport: HTTP, address: DefaultGRPCAddress},
		{name: "JSON", args: []string{"-c", path}, transport: GRPC, address: "json:8081"},
		{name: "flags", args: []string{"-c", path, "-transport=http", "-grpc-address=flag:8081"}, transport: HTTP, address: "flag:8081"},
		{name: "environment", args: []string{"-c", path, "-transport=http", "-grpc-address=flag:8081"}, envTransport: stringPointer(GRPC), envAddress: stringPointer("env:8081"), transport: GRPC, address: "env:8081"},
		{name: "empty environment", args: []string{"-c", path}, envTransport: stringPointer(""), envAddress: stringPointer(""), transport: "", address: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			oldFlags, oldArgs := flag.CommandLine, os.Args
			t.Cleanup(func() { flag.CommandLine, os.Args = oldFlags, oldArgs })
			flag.CommandLine = flag.NewFlagSet("server", flag.ContinueOnError)
			os.Args = append([]string{"server"}, tc.args...)
			for _, key := range []string{"CONFIG", "TRANSPORT", "GRPC_ADDRESS"} {
				t.Setenv(key, "")
				require.NoError(t, os.Unsetenv(key))
			}
			if tc.envTransport != nil {
				t.Setenv("TRANSPORT", *tc.envTransport)
			}
			if tc.envAddress != nil {
				t.Setenv("GRPC_ADDRESS", *tc.envAddress)
			}
			cfg := NewConfig()
			require.Equal(t, tc.transport, cfg.Transport)
			require.Equal(t, tc.address, cfg.GRPCAddress)
		})
	}
	for _, value := range []string{"", "https", "GRPC"} {
		require.Error(t, ValidateTransport(value))
	}
	require.NoError(t, ValidateTransport(HTTP))
	require.NoError(t, ValidateTransport(GRPC))
}
