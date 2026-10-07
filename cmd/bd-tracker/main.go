package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type TrackerTask struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Type        string   `json:"type"`
	Status      string   `json:"status"`
	Priority    int      `json:"priority"`
	Assignee    string   `json:"assignee,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Description string   `json:"description,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
}

type CreateArgs struct {
	Title       string   `json:"title"`
	Type        string   `json:"type,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Priority    int      `json:"priority,omitempty"`
	Description string   `json:"description,omitempty"`
	Parent      string   `json:"parent,omitempty"`
	DependsOn   []string `json:"depends_on,omitempty"`
	FindingOf   string   `json:"finding_of,omitempty"`
}

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      any             `json:"id"`
}

type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
	ID      any             `json:"id"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type Capabilities struct {
	Tools     bool `json:"tools"`
	Providers bool `json:"providers"`
	Tracker   bool `json:"tracker"`
	Commands  bool `json:"commands"`
	Version   int  `json:"version"`
}

type TrackerFrontierResult struct {
	Task *TrackerTask `json:"task,omitempty"`
}

type TrackerClaimParams struct {
	ID string `json:"id"`
}

type TrackerClaimResult struct {
	Task *TrackerTask `json:"task"`
}

type TrackerReadParams struct {
	ID string `json:"id"`
}

type TrackerReadResult struct {
	Task *TrackerTask `json:"task"`
}

type TrackerCloseParams struct {
	ID     string `json:"id"`
	Reason string `json:"reason,omitempty"`
}

type TrackerCloseResult struct {
	Success bool `json:"success"`
}

type TrackerCommentParams struct {
	ID      string `json:"id"`
	Comment string `json:"comment"`
}

type TrackerCommentResult struct {
	Success bool `json:"success"`
}

type TrackerCreateParams struct {
	Title       string   `json:"title"`
	Type        string   `json:"type,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Priority    int      `json:"priority,omitempty"`
	Description string   `json:"description,omitempty"`
	Parent      string   `json:"parent,omitempty"`
	DependsOn   []string `json:"depends_on,omitempty"`
	FindingOf   string   `json:"finding_of,omitempty"`
}

type TrackerCreateResult struct {
	Task *TrackerTask `json:"task"`
}

const (
	MethodCapabilitiesList       = "capabilities/list"
	MethodTrackerFrontier        = "tracker/frontier"
	MethodTrackerClaim           = "tracker/claim"
	MethodTrackerRead            = "tracker/read"
	MethodTrackerClose           = "tracker/close"
	MethodTrackerComment         = "tracker/comment"
	MethodTrackerCreate          = "tracker/create"
	MethodPing                   = "ping"
	MethodShutdown               = "shutdown"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--manifest" {
		printManifest()
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "--capabilities" {
		printCapabilities()
		return
	}

	runPlugin()
}

func printManifest() {
	manifest := map[string]any{
		"name":         "bd-tracker",
		"version":      "1.0.0",
		"description":  "Genie tracker plugin for bd/beads issue tracker",
		"author":       "Genie Team",
		"license":      "MIT",
		"capabilities": []string{"tracker"},
	}
	json.NewEncoder(os.Stdout).Encode(manifest)
}

func printCapabilities() {
	caps := Capabilities{
		Tracker: true,
		Version: 1,
	}
	json.NewEncoder(os.Stdout).Encode(caps)
}

func runPlugin() {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var req Request
		if err := decoder.Decode(&req); err != nil {
			return
		}

		var resp Response
		resp.JSONRPC = "2.0"
		resp.ID = req.ID

		switch req.Method {
		case MethodCapabilitiesList:
			caps := Capabilities{Tracker: true, Version: 1}
			resp.Result, _ = json.Marshal(caps)

		case MethodTrackerFrontier:
			resp.Result = handleFrontier(req.Params)

		case MethodTrackerClaim:
			resp.Result = handleClaim(req.Params)

		case MethodTrackerRead:
			resp.Result = handleRead(req.Params)

		case MethodTrackerClose:
			resp.Result = handleClose(req.Params)

		case MethodTrackerComment:
			resp.Result = handleComment(req.Params)

		case MethodTrackerCreate:
			resp.Result = handleCreate(req.Params)

		case MethodPing:
			resp.Result, _ = json.Marshal(map[string]string{"status": "ok"})

		case MethodShutdown:
			resp.Result, _ = json.Marshal(map[string]string{"status": "shutting down"})
			encoder.Encode(resp)
			return

		default:
			resp.Error = &Error{Code: -32601, Message: "Method not found: " + req.Method}
		}

		encoder.Encode(resp)
	}
}

func handleFrontier(params json.RawMessage) json.RawMessage {
	ctx := context.Background()
	task, err := frontier(ctx)
	if err != nil {
		return errorResult(-32603, err.Error())
	}
	result := TrackerFrontierResult{Task: task}
	data, _ := json.Marshal(result)
	return data
}

func handleClaim(params json.RawMessage) json.RawMessage {
	var p TrackerClaimParams
	if err := json.Unmarshal(params, &p); err != nil {
		return errorResult(-32602, "invalid params: "+err.Error())
	}
	ctx := context.Background()
	task, err := claim(ctx, p.ID)
	if err != nil {
		return errorResult(-32603, err.Error())
	}
	result := TrackerClaimResult{Task: task}
	data, _ := json.Marshal(result)
	return data
}

func handleRead(params json.RawMessage) json.RawMessage {
	var p TrackerReadParams
	if err := json.Unmarshal(params, &p); err != nil {
		return errorResult(-32602, "invalid params: "+err.Error())
	}
	ctx := context.Background()
	task, err := read(ctx, p.ID)
	if err != nil {
		return errorResult(-32603, err.Error())
	}
	result := TrackerReadResult{Task: task}
	data, _ := json.Marshal(result)
	return data
}

func handleClose(params json.RawMessage) json.RawMessage {
	var p TrackerCloseParams
	if err := json.Unmarshal(params, &p); err != nil {
		return errorResult(-32602, "invalid params: "+err.Error())
	}
	ctx := context.Background()
	err := closeTask(ctx, p.ID, p.Reason)
	if err != nil {
		return errorResult(-32603, err.Error())
	}
	result := TrackerCloseResult{Success: true}
	data, _ := json.Marshal(result)
	return data
}

func handleComment(params json.RawMessage) json.RawMessage {
	var p TrackerCommentParams
	if err := json.Unmarshal(params, &p); err != nil {
		return errorResult(-32602, "invalid params: "+err.Error())
	}
	ctx := context.Background()
	err := comment(ctx, p.ID, p.Comment)
	if err != nil {
		return errorResult(-32603, err.Error())
	}
	result := TrackerCommentResult{Success: true}
	data, _ := json.Marshal(result)
	return data
}

func handleCreate(params json.RawMessage) json.RawMessage {
	var p TrackerCreateParams
	if err := json.Unmarshal(params, &p); err != nil {
		return errorResult(-32602, "invalid params: "+err.Error())
	}
	ctx := context.Background()
	args := CreateArgs{
		Title:       p.Title,
		Type:        p.Type,
		Labels:      p.Labels,
		Priority:    p.Priority,
		Description: p.Description,
		Parent:      p.Parent,
		DependsOn:   p.DependsOn,
		FindingOf:   p.FindingOf,
	}
	task, err := create(ctx, args)
	if err != nil {
		return errorResult(-32603, err.Error())
	}
	result := TrackerCreateResult{Task: task}
	data, _ := json.Marshal(result)
	return data
}

func errorResult(code int, message string) json.RawMessage {
	resp := Response{
		JSONRPC: "2.0",
		Error:   &Error{Code: code, Message: message},
	}
	data, _ := json.Marshal(resp)
	return data
}

func runBD(ctx context.Context, binary string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, binary, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("bd %s failed: %w: %s", strings.Join(args, " "), err, string(output))
	}
	return string(output), nil
}

func frontier(ctx context.Context) (*TrackerTask, error) {
	binary := getBinary()
	output, err := runBD(ctx, binary, "ready")
	if err != nil {
		return nil, fmt.Errorf("bd ready failed: %w", err)
	}
	task, err := parseFrontier(output)
	if err != nil {
		return nil, fmt.Errorf("parse frontier: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("no ready tasks available")
	}
	return task, nil
}

func claim(ctx context.Context, id string) (*TrackerTask, error) {
	binary := getBinary()
	_, err := runBD(ctx, binary, "update", id, "--claim")
	if err != nil {
		return nil, fmt.Errorf("bd claim failed: %w", err)
	}
	return read(ctx, id)
}

func read(ctx context.Context, id string) (*TrackerTask, error) {
	binary := getBinary()
	_, err := runBD(ctx, binary, "show", id)
	if err != nil {
		return nil, fmt.Errorf("bd show failed: %w", err)
	}
	return parseShow(id, "")
}

func closeTask(ctx context.Context, id string, reason string) error {
	binary := getBinary()
	args := []string{"close", id}
	if reason != "" {
		args = append(args, "--reason", reason)
	}
	_, err := runBD(ctx, binary, args...)
	return err
}

func comment(ctx context.Context, id string, comment string) error {
	binary := getBinary()
	_, err := runBD(ctx, binary, "comment", id, "--message", comment)
	return err
}

func create(ctx context.Context, args CreateArgs) (*TrackerTask, error) {
	binary := getBinary()
	bdArgs := []string{"create", "--title", args.Title, "--type", args.Type}

	if args.Priority > 0 {
		bdArgs = append(bdArgs, "--priority", fmt.Sprintf("%d", args.Priority))
	}
	if args.Description != "" {
		bdArgs = append(bdArgs, "--description", args.Description)
	}
	if len(args.Labels) > 0 {
		bdArgs = append(bdArgs, "--labels", strings.Join(args.Labels, ","))
	}
	if args.Parent != "" {
		bdArgs = append(bdArgs, "--parent", args.Parent)
	}
	if len(args.DependsOn) > 0 {
		bdArgs = append(bdArgs, "--depends-on", strings.Join(args.DependsOn, ","))
	}
	if args.FindingOf != "" {
		bdArgs = append(bdArgs, "--finding-of", args.FindingOf)
	}

	output, err := runBD(ctx, binary, bdArgs...)
	if err != nil {
		return nil, fmt.Errorf("bd create failed: %w", err)
	}

	// Parse the output to get the created task ID
	// bd create output format: "Created og-xxx"
	re := regexp.MustCompile(`Created\s+(\S+)`)
	matches := re.FindStringSubmatch(output)
	if matches == nil {
		return nil, fmt.Errorf("could not parse created task ID from output: %s", output)
	}

	taskID := matches[1]
	return read(ctx, taskID)
}

func getBinary() string {
	if bin := os.Getenv("BD_TRACKER_BINARY"); bin != "" {
		return bin
	}
	return "bd"
}

func parseFrontier(output string) (*TrackerTask, error) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Status:") || strings.HasPrefix(line, "---") {
			continue
		}
		re := regexp.MustCompile(`^[○◐●✓❄]\s+(\S+)\s+●\s+(\S+)\s+(\S+)\s+(.+)$`)
		matches := re.FindStringSubmatch(line)
		if matches != nil {
			id := matches[1]
			status := parseStatus(matches[2])
			priority := parsePriority(matches[3])
			rest := matches[4]
			taskType, title := extractTypeAndTitle(rest)
			return &TrackerTask{
				ID:        id,
				Title:     title,
				Type:      taskType,
				Status:    status,
				Priority:  priority,
				CreatedAt: time.Now().Format(time.RFC3339),
				UpdatedAt: time.Now().Format(time.RFC3339),
			}, nil
		}
	}
	return nil, fmt.Errorf("no ready tasks available")
}

func parseShow(id, output string) (*TrackerTask, error) {
	binary := getBinary()
	output, err := runBD(context.Background(), binary, "show", id)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(output, "\n")
	var task *TrackerTask

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "○ ") || strings.HasPrefix(line, "◐ ") || strings.HasPrefix(line, "● ") || strings.HasPrefix(line, "✓ ") || strings.HasPrefix(line, "❄ ") {
			re := regexp.MustCompile(`^[○◐●✓❄]\s+(\S+)\s+(\[?\S+\]?)\s*[·:]\s*(.+)$`)
			matches := re.FindStringSubmatch(line)
			if matches != nil {
				desc := matches[3]
				taskType, title := extractTypeAndTitle(desc)
				if task == nil {
					task = &TrackerTask{
						ID:        id,
						Title:     title,
						Type:      taskType,
						Status:    "open",
						Priority:  2,
						CreatedAt: time.Now().Format(time.RFC3339),
						UpdatedAt: time.Now().Format(time.RFC3339),
					}
				}
			}
		}
		if strings.HasPrefix(line, "Owner:") {
			if task != nil && strings.Contains(line, "Assignee:") {
				parts := strings.Split(line, "·")
				for _, part := range parts {
					part = strings.TrimSpace(part)
					if strings.HasPrefix(part, "Assignee:") {
						task.Assignee = strings.TrimSpace(strings.TrimPrefix(part, "Assignee:"))
					}
				}
			}
		}
		if strings.HasPrefix(line, "Created:") {
			parts := strings.Split(line, "·")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "Created:") {
					if t, err := time.Parse("2006-01-02", strings.TrimSpace(strings.TrimPrefix(part, "Created:"))); err == nil {
						if task != nil {
							task.CreatedAt = t.Format(time.RFC3339)
						}
					}
				}
				if strings.HasPrefix(part, "Updated:") {
					if t, err := time.Parse("2006-01-02", strings.TrimSpace(strings.TrimPrefix(part, "Updated:"))); err == nil {
						if task != nil {
							task.UpdatedAt = t.Format(time.RFC3339)
						}
					}
				}
			}
		}
		if strings.HasPrefix(line, "Labels:") {
			if task != nil {
				labelsStr := strings.TrimSpace(strings.TrimPrefix(line, "Labels:"))
				task.Labels = strings.Split(labelsStr, ", ")
			}
		}
	}

	if task == nil {
		return nil, fmt.Errorf("could not parse bd show output")
	}
	return task, nil
}

func parseStatus(status string) string {
	switch status {
	case "P0", "0":
		return "critical"
	case "P1", "1":
		return "high"
	case "P2", "2":
		return "medium"
	case "P3", "3":
		return "low"
	case "P4", "4":
		return "backlog"
	default:
		return "unknown"
	}
}

func parsePriority(priority string) int {
	switch priority {
	case "P0", "0":
		return 0
	case "P1", "1":
		return 1
	case "P2", "2":
		return 2
	case "P3", "3":
		return 3
	case "P4", "4":
		return 4
	default:
		return 2
	}
}

func extractTypeAndTitle(desc string) (string, string) {
	parts := strings.Split(desc, "·")
	if len(parts) >= 2 {
		typePart := strings.TrimSpace(parts[0])
		typePart = strings.Trim(typePart, "[]")
		titlePart := strings.TrimSpace(parts[1])
		if idx := strings.Index(titlePart, ":"); idx >= 0 {
			titlePart = strings.TrimSpace(titlePart[idx+1:])
		}
		return typePart, titlePart
	}
	if idx := strings.Index(desc, ":"); idx >= 0 {
		return "task", strings.TrimSpace(desc[idx+1:])
	}
	return "task", desc
}