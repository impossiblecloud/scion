// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"os"
	"testing"

	"github.com/ptone/scion-agent/pkg/hubclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatherAndSubmitEnv_NonInteractiveRejectsNeededKeys(t *testing.T) {
	// Save and restore package-level state
	origNonInteractive := nonInteractive
	origAutoConfirm := autoConfirm
	origOutputFormat := outputFormat
	defer func() {
		nonInteractive = origNonInteractive
		autoConfirm = origAutoConfirm
		outputFormat = origOutputFormat
	}()

	nonInteractive = true
	autoConfirm = true // --non-interactive implies --yes

	// Even though the key is available in the local env, non-interactive
	// mode should refuse to gather and send it.
	os.Setenv("TEST_SECRET_KEY", "secret-value")
	defer os.Unsetenv("TEST_SECRET_KEY")

	resp := &hubclient.CreateAgentResponse{
		Agent: &hubclient.Agent{ID: "agent-1"},
		EnvGather: &hubclient.EnvGatherResponse{
			AgentID:  "agent-1",
			Required: []string{"TEST_SECRET_KEY"},
			Needs:    []string{"TEST_SECRET_KEY"},
		},
	}

	_, err := gatherAndSubmitEnv(context.Background(), nil, "grove-1", resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
	assert.Contains(t, err.Error(), "TEST_SECRET_KEY")
}

func TestGatherAndSubmitEnv_NonInteractiveAllowsWhenAllSatisfied(t *testing.T) {
	// Save and restore package-level state
	origNonInteractive := nonInteractive
	origAutoConfirm := autoConfirm
	defer func() {
		nonInteractive = origNonInteractive
		autoConfirm = origAutoConfirm
	}()

	nonInteractive = true
	autoConfirm = true

	// When Needs is empty (Hub/Broker satisfied everything), non-interactive
	// should succeed.
	resp := &hubclient.CreateAgentResponse{
		Agent: &hubclient.Agent{ID: "agent-2"},
		EnvGather: &hubclient.EnvGatherResponse{
			AgentID:  "agent-2",
			Required: []string{"GEMINI_API_KEY"},
			HubHas:   []hubclient.EnvSource{{Key: "GEMINI_API_KEY", Scope: "grove"}},
			Needs:    []string{},
		},
	}

	result, err := gatherAndSubmitEnv(context.Background(), nil, "grove-1", resp)
	require.NoError(t, err)
	// Should return the original response since no env was gathered
	assert.Equal(t, resp, result)
}

func TestGatherAndSubmitEnv_NonInteractiveMultipleKeys(t *testing.T) {
	// Save and restore package-level state
	origNonInteractive := nonInteractive
	origAutoConfirm := autoConfirm
	origOutputFormat := outputFormat
	defer func() {
		nonInteractive = origNonInteractive
		autoConfirm = origAutoConfirm
		outputFormat = origOutputFormat
	}()

	nonInteractive = true
	autoConfirm = true
	outputFormat = "json" // Suppress stderr output for cleaner test

	resp := &hubclient.CreateAgentResponse{
		Agent: &hubclient.Agent{ID: "agent-3"},
		EnvGather: &hubclient.EnvGatherResponse{
			AgentID:  "agent-3",
			Required: []string{"KEY_A", "KEY_B"},
			Needs:    []string{"KEY_A", "KEY_B"},
		},
	}

	_, err := gatherAndSubmitEnv(context.Background(), nil, "grove-1", resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
	assert.Contains(t, err.Error(), "KEY_A")
	assert.Contains(t, err.Error(), "KEY_B")
}
