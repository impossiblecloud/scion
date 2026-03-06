package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainerBridgeEndpoint(t *testing.T) {
	tests := []struct {
		name        string
		hubEndpoint string
		runtimeName string
		want        string
	}{
		{
			name:        "podman with localhost",
			hubEndpoint: "http://localhost:8080",
			runtimeName: "podman",
			want:        "http://host.containers.internal:8080",
		},
		{
			name:        "docker with localhost",
			hubEndpoint: "http://localhost:8080",
			runtimeName: "docker",
			want:        "http://host.docker.internal:8080",
		},
		{
			name:        "podman with 127.0.0.1",
			hubEndpoint: "http://127.0.0.1:9090",
			runtimeName: "podman",
			want:        "http://host.containers.internal:9090",
		},
		{
			name:        "docker with ipv6 loopback",
			hubEndpoint: "http://[::1]:8080",
			runtimeName: "docker",
			want:        "http://host.docker.internal:8080",
		},
		{
			name:        "non-localhost endpoint unchanged",
			hubEndpoint: "https://hub.example.com:443",
			runtimeName: "podman",
			want:        "",
		},
		{
			name:        "kubernetes returns empty",
			hubEndpoint: "http://localhost:8080",
			runtimeName: "kubernetes",
			want:        "",
		},
		{
			name:        "apple runtime returns empty",
			hubEndpoint: "http://localhost:8080",
			runtimeName: "apple-container",
			want:        "",
		},
		{
			name:        "empty runtime returns empty",
			hubEndpoint: "http://localhost:8080",
			runtimeName: "",
			want:        "",
		},
		{
			name:        "empty endpoint returns empty",
			hubEndpoint: "",
			runtimeName: "podman",
			want:        "",
		},
		{
			name:        "invalid URL returns empty",
			hubEndpoint: "://not-a-url",
			runtimeName: "podman",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containerBridgeEndpoint(tt.hubEndpoint, tt.runtimeName)
			assert.Equal(t, tt.want, got)
		})
	}
}
