package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := server.NewMCPServer(
		"AWS Security Scanner",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	scannerTool := mcp.NewTool("scanner",
		mcp.WithDescription("Scans text looking for AWS API keys and reporting any potential issues."),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("The text you want to be scanned"),
			mcp.MaxLength(5000),
		),
	)

	s.AddTool(scannerTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		content := request.Params.Arguments["content"].(string)

		results := SecurityScan(content, securityChecksList)
		if len(results) == 0 {
			return mcp.NewToolResultText("No secrets were found."), nil
		}

		sb := strings.Builder{}
		for _, r := range results {
			sb.WriteString("WARNING ")
			sb.WriteString(r.Name)
			sb.WriteString(" found in content!\n")
		}

		return mcp.NewToolResultText(sb.String()), nil
	})

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

type SecurityChecks struct {
	Name        string
	Description string
	Type        string
	Value       string
	Regex       *regexp.Regexp
	Locations   [][]int
}

var securityChecksList = []SecurityChecks{
	{
		Name:        "AWS API Key",
		Description: "",
		Type:        "regex",
		Regex:       regexp.MustCompile("((?:A3T[A-Z0-9]|AKIA|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[A-Z0-9]{16})")},
	{
		Name:        "AWS Secret Key",
		Description: "",
		Type:        "regex",
		Regex:       regexp.MustCompile(`(?i)aws(.{0,20})?(?-i)['"][0-9a-zA-Z/+]{40}['"]`)},
	{
		Name:        "Amazon MWS Auth Token",
		Description: "", Type: "regex",
		Regex: regexp.MustCompile(`amzn\.mws\.[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)},
	{
		Name:        "AWS AppSync GraphQL Key",
		Description: "",
		Type:        "regex",
		Regex:       regexp.MustCompile("da2-[a-z0-9]{26}")},
}

func SecurityScan(content string, securityChecks []SecurityChecks) []SecurityChecks {
	results := []SecurityChecks{}

	for _, check := range securityChecks {
		res := check.Regex.FindAllIndex([]byte(content), -1)
		if len(res) != 0 {
			check.Locations = res
			results = append(results, check)
		}
	}

	return results
}
