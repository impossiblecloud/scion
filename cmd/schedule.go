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
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/GoogleCloudPlatform/scion/pkg/hubclient"
	"github.com/spf13/cobra"
)

// Schedule command flags
var (
	scheduleType      string
	scheduleStatus    string
	scheduleIn        string
	scheduleAt        string
	scheduleAgent     string
	scheduleMessage   string
	scheduleInterrupt bool
)

// scheduleCmd is the top-level command group for schedule management.
var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Manage scheduled events",
	Long:  `List, inspect, create, and cancel scheduled events for the current grove.`,
}

// scheduleListCmd lists scheduled events for the current grove.
var scheduleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List scheduled events",
	RunE:  runScheduleList,
}

// scheduleGetCmd shows details of a specific scheduled event.
var scheduleGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get details of a scheduled event",
	Args:  cobra.ExactArgs(1),
	RunE:  runScheduleGet,
}

// scheduleCancelCmd cancels a pending scheduled event.
var scheduleCancelCmd = &cobra.Command{
	Use:   "cancel <id>",
	Short: "Cancel a pending scheduled event",
	Args:  cobra.ExactArgs(1),
	RunE:  runScheduleCancel,
}

// scheduleCreateCmd creates a new one-shot scheduled event.
var scheduleCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a one-shot scheduled event",
	Long: `Create a one-shot scheduled event. Requires --type, timing (--in or --at),
and type-specific flags (e.g. --agent and --message for message events).`,
	RunE: runScheduleCreate,
}

func runScheduleList(cmd *cobra.Command, args []string) error {
	hubCtx, err := CheckHubAvailabilityWithOptions(grovePath, true)
	if err != nil {
		return err
	}
	if hubCtx == nil {
		return fmt.Errorf("scheduled events require Hub mode (use 'scion hub enable' first)")
	}

	if !isJSONOutput() {
		PrintUsingHub(hubCtx.Endpoint)
	}

	groveID, err := GetGroveID(hubCtx)
	if err != nil {
		return wrapHubError(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := &hubclient.ListScheduledEventsOptions{}
	if scheduleStatus != "" {
		opts.Status = scheduleStatus
	}
	if scheduleType != "" {
		opts.EventType = scheduleType
	}

	resp, err := hubCtx.Client.ScheduledEvents(groveID).List(ctx, opts)
	if err != nil {
		return wrapHubError(fmt.Errorf("failed to list scheduled events: %w", err))
	}

	if isJSONOutput() {
		return outputJSON(resp)
	}

	events := resp.Events
	if len(events) == 0 {
		fmt.Println("No scheduled events found.")
		return nil
	}

	fmt.Printf("SCHEDULED EVENTS (%d)\n", len(events))
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTYPE\tSTATUS\tFIRE AT\tCREATED")
	for _, evt := range events {
		id := evt.ID
		if len(id) > 8 {
			id = id[:8]
		}
		fireAt := formatScheduleTime(evt.FireAt, evt.Status)
		created := formatRelativeTime(evt.CreatedAt)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", id, evt.EventType, evt.Status, fireAt, created)
	}
	w.Flush()

	return nil
}

func runScheduleGet(cmd *cobra.Command, args []string) error {
	eventID := args[0]

	hubCtx, err := CheckHubAvailabilityWithOptions(grovePath, true)
	if err != nil {
		return err
	}
	if hubCtx == nil {
		return fmt.Errorf("scheduled events require Hub mode (use 'scion hub enable' first)")
	}

	if !isJSONOutput() {
		PrintUsingHub(hubCtx.Endpoint)
	}

	groveID, err := GetGroveID(hubCtx)
	if err != nil {
		return wrapHubError(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	evt, err := hubCtx.Client.ScheduledEvents(groveID).Get(ctx, eventID)
	if err != nil {
		return wrapHubError(fmt.Errorf("failed to get scheduled event: %w", err))
	}

	if isJSONOutput() {
		return outputJSON(evt)
	}

	fmt.Printf("Scheduled Event: %s\n", evt.ID)
	fmt.Printf("  Type:       %s\n", evt.EventType)
	fmt.Printf("  Status:     %s\n", evt.Status)
	fmt.Printf("  Fire At:    %s (%s)\n", evt.FireAt.Format(time.RFC3339), formatScheduleTime(evt.FireAt, evt.Status))
	fmt.Printf("  Grove:      %s\n", evt.GroveID)
	fmt.Printf("  Created:    %s\n", evt.CreatedAt.Format(time.RFC3339))
	if evt.CreatedBy != "" {
		fmt.Printf("  Created By: %s\n", evt.CreatedBy)
	}
	if evt.FiredAt != nil {
		fmt.Printf("  Fired At:   %s\n", evt.FiredAt.Format(time.RFC3339))
	}
	if evt.Error != "" {
		fmt.Printf("  Error:      %s\n", evt.Error)
	}

	// Parse and display payload details
	if evt.Payload != "" {
		var payload map[string]interface{}
		if json.Unmarshal([]byte(evt.Payload), &payload) == nil {
			if agentName, ok := payload["agentName"].(string); ok && agentName != "" {
				fmt.Printf("  Agent:      %s\n", agentName)
			}
			if message, ok := payload["message"].(string); ok && message != "" {
				fmt.Printf("  Message:    %q\n", message)
			}
		}
	}

	return nil
}

func runScheduleCancel(cmd *cobra.Command, args []string) error {
	eventID := args[0]

	hubCtx, err := CheckHubAvailabilityWithOptions(grovePath, true)
	if err != nil {
		return err
	}
	if hubCtx == nil {
		return fmt.Errorf("scheduled events require Hub mode (use 'scion hub enable' first)")
	}

	if !isJSONOutput() {
		PrintUsingHub(hubCtx.Endpoint)
	}

	groveID, err := GetGroveID(hubCtx)
	if err != nil {
		return wrapHubError(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := hubCtx.Client.ScheduledEvents(groveID).Cancel(ctx, eventID); err != nil {
		return wrapHubError(fmt.Errorf("failed to cancel scheduled event: %w", err))
	}

	return outputActionResult(ActionResult{
		Status:  "ok",
		Command: "schedule cancel",
		Message: fmt.Sprintf("Scheduled event %s cancelled.", eventID),
	})
}

func runScheduleCreate(cmd *cobra.Command, args []string) error {
	if scheduleType == "" {
		return fmt.Errorf("--type is required")
	}
	if scheduleIn == "" && scheduleAt == "" {
		return fmt.Errorf("either --in or --at is required")
	}
	if scheduleIn != "" && scheduleAt != "" {
		return fmt.Errorf("--in and --at are mutually exclusive")
	}

	// Validate type-specific flags
	switch scheduleType {
	case "message":
		if scheduleAgent == "" {
			return fmt.Errorf("--agent is required for message events")
		}
		if scheduleMessage == "" {
			return fmt.Errorf("--message is required for message events")
		}
	default:
		return fmt.Errorf("unsupported event type: %q (supported: message)", scheduleType)
	}

	hubCtx, err := CheckHubAvailabilityWithOptions(grovePath, true)
	if err != nil {
		return err
	}
	if hubCtx == nil {
		return fmt.Errorf("scheduled events require Hub mode (use 'scion hub enable' first)")
	}

	if !isJSONOutput() {
		PrintUsingHub(hubCtx.Endpoint)
	}

	groveID, err := GetGroveID(hubCtx)
	if err != nil {
		return wrapHubError(err)
	}

	req := &hubclient.CreateScheduledEventRequest{
		EventType: scheduleType,
		AgentName: scheduleAgent,
		Message:   scheduleMessage,
		Interrupt: scheduleInterrupt,
	}

	if scheduleIn != "" {
		req.FireIn = scheduleIn
	} else {
		req.FireAt = scheduleAt
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	evt, err := hubCtx.Client.ScheduledEvents(groveID).Create(ctx, req)
	if err != nil {
		return wrapHubError(fmt.Errorf("failed to create scheduled event: %w", err))
	}

	if isJSONOutput() {
		return outputJSON(evt)
	}

	fmt.Printf("Scheduled event created: %s\n", evt.ID)
	fmt.Printf("  Type:    %s\n", evt.EventType)
	fmt.Printf("  Fire At: %s\n", evt.FireAt.Format(time.RFC3339))

	return nil
}

// formatScheduleTime returns a human-readable time description for an event.
func formatScheduleTime(t time.Time, status string) string {
	if status == "pending" {
		diff := time.Until(t)
		if diff <= 0 {
			return "now"
		}
		return "in " + formatScheduleDuration(diff)
	}
	// Reuse formatRelativeTime from hub.go for past times
	return formatRelativeTime(t)
}

// formatScheduleDuration returns a human-readable duration string.
func formatScheduleDuration(d time.Duration) string {
	if d < time.Minute {
		s := int(d.Seconds())
		if s <= 1 {
			return "1 second"
		}
		return fmt.Sprintf("%d seconds", s)
	}
	if d < time.Hour {
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", m)
	}
	if d < 24*time.Hour {
		h := int(d.Hours())
		if h == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", h)
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}

func init() {
	rootCmd.AddCommand(scheduleCmd)

	scheduleCmd.AddCommand(scheduleListCmd)
	scheduleCmd.AddCommand(scheduleGetCmd)
	scheduleCmd.AddCommand(scheduleCancelCmd)
	scheduleCmd.AddCommand(scheduleCreateCmd)

	// List flags
	scheduleListCmd.Flags().StringVar(&scheduleStatus, "status", "", "Filter by status (pending, fired, cancelled, expired)")
	scheduleListCmd.Flags().StringVar(&scheduleType, "type", "", "Filter by event type (e.g. message)")

	// Create flags
	scheduleCreateCmd.Flags().StringVar(&scheduleType, "type", "", "Event type (required, e.g. message)")
	scheduleCreateCmd.Flags().StringVar(&scheduleIn, "in", "", "Schedule after a duration (e.g. 30m, 1h)")
	scheduleCreateCmd.Flags().StringVar(&scheduleAt, "at", "", "Schedule at an absolute time (ISO 8601)")
	scheduleCreateCmd.Flags().StringVar(&scheduleAgent, "agent", "", "Target agent name (for message events)")
	scheduleCreateCmd.Flags().StringVar(&scheduleMessage, "message", "", "Message body (for message events)")
	scheduleCreateCmd.Flags().BoolVar(&scheduleInterrupt, "interrupt", false, "Interrupt the agent (for message events)")
}
