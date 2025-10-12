package applecalendar

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const defaultAppleCalendarPath = "/opt/homebrew/bin/apple-calendar"

// Client represents an apple-calendar command client
type Client struct {
	executablePath string
	ctx            context.Context
	cancel         context.CancelFunc
}

// New creates a new apple-calendar client with default executable path
func New() *Client {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	return &Client{
		executablePath: defaultAppleCalendarPath,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Close cancels the context and releases resources
func (c *Client) Close() {
	if c.cancel != nil {
		c.cancel()
	}
}

// GetCalendarList executes apple-calendar list-calendars --format=json and returns calendar names
func (c *Client) GetCalendarList() ([]string, error) {
	output, err := c.executeCommand("list-calendars", "--format=json")
	if err != nil {
		return nil, fmt.Errorf("failed to execute apple-calendar list-calendars: %w", err)
	}

	// Parse JSON output - assuming it returns an array of calendar objects
	var calendars []map[string]interface{}
	if err := json.Unmarshal(output, &calendars); err != nil {
		return nil, fmt.Errorf("failed to parse calendar list JSON: %w", err)
	}

	var names []string
	for _, cal := range calendars {
		if name, ok := cal["title"].(string); ok {
			names = append(names, name)
		}
	}

	// Fallback to default calendars if none found
	if len(names) == 0 {
		names = []string{"Work", "Personal"}
	}

	return names, nil
}

// GetEventList executes apple-calendar list-events command with given calendars and params, returns raw output
func (c *Client) GetEventList(calendars []string, params map[string]string) ([]byte, error) {
	// Build command: apple-calendar list-events --calendars=foo,bar --format=json [additional args...]
	cmdArgs := []string{"list-events"}

	// Add calendars as --calendars=foo,bar format
	if len(calendars) > 0 {
		calendarsArg := "--calendars=" + strings.Join(calendars, ",")
		cmdArgs = append(cmdArgs, calendarsArg)
	}

	cmdArgs = append(cmdArgs, "--format=json")

	// Convert map[string]string to command line arguments
	boolParams := map[string]bool{
		"exclude-all-day":    true,
		"exclude-long-event": true,
	}

	for key, value := range params {
		if isBoolParam, exists := boolParams[key]; exists && isBoolParam {
			// For boolean parameters, only add the flag if value is "true"
			if value == "true" {
				cmdArgs = append(cmdArgs, "--"+key)
			}
			// Skip if value is "false" or anything else
		} else {
			// For non-boolean parameters, add as --key=value
			cmdArgs = append(cmdArgs, "--"+key+"="+value)
		}
	}

	return c.executeCommand(cmdArgs...)
}

// executeCommand executes apple-calendar command with given arguments (private method)
func (c *Client) executeCommand(args ...string) ([]byte, error) {
	cmd := exec.CommandContext(c.ctx, c.executablePath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute apple-calendar %v: %w", args, err)
	}
	return output, nil
}
