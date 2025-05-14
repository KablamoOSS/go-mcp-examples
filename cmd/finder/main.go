package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := server.NewMCPServer(
		"Filesystem Finder",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	scannerTool := mcp.NewTool("finder",
		mcp.WithDescription("Finds the name of a file or directory on the local filesystem if supplied the name and path"),
		mcp.WithString("filename",
			mcp.Required(),
			mcp.Description("Name of the file you are looking for"),
			mcp.MaxLength(100),
		),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("File path you want to start looking into"),
			mcp.MaxLength(200),
		),
	)

	s.AddTool(scannerTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filename := request.Params.Arguments["filename"].(string)
		path := request.Params.Arguments["path"].(string)

		sb := strings.Builder{}
		matches := 0
		err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if matches >= 100 {
				return errors.New("too many matches")
			}

			if strings.Contains(p, filename) {
				sb.WriteString("Found matching path: ")
				sb.WriteString(p)
				sb.WriteString("\n")
			}
			return nil
		})

		if err != nil {
			sb.WriteString(err.Error())
		}

		return mcp.NewToolResultText(sb.String()), nil
	})

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
