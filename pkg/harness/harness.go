package harness

import (
	"github.com/ptone/scion-agent/pkg/api"
)

// Harness is now defined in pkg/api to avoid import cycles
type Harness = api.Harness

func New(provider string) Harness {
	switch provider {
	case "claude":
		return &ClaudeCode{}
	case "gemini":
		return &GeminiCLI{}
	default:
		return &Generic{}
	}
}