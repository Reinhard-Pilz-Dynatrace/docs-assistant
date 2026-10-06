package github

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func NewFromEnvironment() (*Client, error) {
	repository := os.Getenv("GITHUB_REPOSITORY")
	parts := strings.SplitN(repository, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("GITHUB_REPOSITORY must be set as owner/name")
	}
	token := os.Getenv("GH_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		output, err := exec.Command("gh", "auth", "token").Output()
		if err != nil {
			return nil, fmt.Errorf("set GH_TOKEN or authenticate with gh: %w", err)
		}
		token = strings.TrimSpace(string(output))
	}
	if token == "" {
		return nil, fmt.Errorf("GitHub token is empty")
	}
	return &Client{
		Owner:      parts[0],
		Repository: parts[1],
		Token:      token,
		APIBase:    os.Getenv("GITHUB_API_URL"),
	}, nil
}
