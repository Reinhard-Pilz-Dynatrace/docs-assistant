package github

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiBase = "https://api.github.com"

var workflowLabels = map[string]struct{}{
	"vi:in-progress":         {},
	"vi:ready-for-docs":      {},
	"vi:needs-clarification": {},
	"vi:docs-review":         {},
	"vi:done":                {},
}

type Client struct {
	Owner      string
	Repository string
	Token      string
	APIBase    string
	HTTPClient *http.Client
}

type Label struct {
	Name string `json:"name"`
}

type Issue struct {
	Number  int     `json:"number"`
	Title   string  `json:"title"`
	Body    string  `json:"body"`
	HTMLURL string  `json:"html_url"`
	State   string  `json:"state"`
	Labels  []Label `json:"labels"`
}

type PullRequest struct {
	Number         int        `json:"number"`
	Title          string     `json:"title"`
	Body           string     `json:"body"`
	HTMLURL        string     `json:"html_url"`
	State          string     `json:"state"`
	MergedAt       *time.Time `json:"merged_at"`
	MergeCommitSHA string     `json:"merge_commit_sha"`
	Base           struct {
		Ref string `json:"ref"`
	} `json:"base"`
}

type PullRequestFile struct {
	Filename string `json:"filename"`
	Status   string `json:"status"`
	Patch    string `json:"patch"`
}

type Repository struct {
	DefaultBranch string `json:"default_branch"`
}

type FileContent struct {
	Path     string `json:"path"`
	SHA      string `json:"sha"`
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

type CreatedPullRequest struct {
	Number  int    `json:"number"`
	HTMLURL string `json:"html_url"`
}

type apiError struct {
	Message string `json:"message"`
}

func (client *Client) base() string {
	if client.APIBase != "" {
		return strings.TrimSuffix(client.APIBase, "/")
	}
	return apiBase
}

func (client *Client) httpClient() *http.Client {
	if client.HTTPClient != nil {
		return client.HTTPClient
	}
	return http.DefaultClient
}

func (client *Client) request(ctx context.Context, method, path string, input, output any) (int, error) {
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return 0, fmt.Errorf("encode GitHub request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, client.base()+path, body)
	if err != nil {
		return 0, fmt.Errorf("create GitHub request: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Set("Content-Type", "application/json")
	if client.Token != "" {
		request.Header.Set("Authorization", "Bearer "+client.Token)
	}
	response, err := client.httpClient().Do(request)
	if err != nil {
		return 0, fmt.Errorf("call GitHub API: %w", err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		return response.StatusCode, fmt.Errorf("read GitHub response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var detail apiError
		_ = json.Unmarshal(data, &detail)
		return response.StatusCode, fmt.Errorf("GitHub API %s %s: %s", method, path, detail.Message)
	}
	if output != nil && len(data) > 0 {
		if err := json.Unmarshal(data, output); err != nil {
			return response.StatusCode, fmt.Errorf("decode GitHub response: %w", err)
		}
	}
	return response.StatusCode, nil
}

func (client *Client) GetIssue(ctx context.Context, number int) (Issue, error) {
	var issue Issue
	_, err := client.request(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/issues/%d", client.Owner, client.Repository, number), nil, &issue)
	return issue, err
}

func (client *Client) GetPullRequest(ctx context.Context, number int) (PullRequest, error) {
	var pull PullRequest
	_, err := client.request(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/%d", client.Owner, client.Repository, number), nil, &pull)
	return pull, err
}

func (client *Client) GetPullRequestFiles(ctx context.Context, number int) ([]PullRequestFile, error) {
	var files []PullRequestFile
	_, err := client.request(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/%d/files?per_page=100", client.Owner, client.Repository, number), nil, &files)
	return files, err
}

func (client *Client) GetFile(ctx context.Context, path, ref string) ([]byte, string, error) {
	query := url.Values{"ref": []string{ref}}
	var file FileContent
	endpoint := fmt.Sprintf("/repos/%s/%s/contents/%s?%s", client.Owner, client.Repository, path, query.Encode())
	if _, err := client.request(ctx, http.MethodGet, endpoint, nil, &file); err != nil {
		return nil, "", err
	}
	if file.Encoding != "base64" {
		return nil, "", fmt.Errorf("unsupported GitHub file encoding %q for %s", file.Encoding, path)
	}
	content, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(file.Content, "\n", ""))
	if err != nil {
		return nil, "", fmt.Errorf("decode GitHub file %s: %w", path, err)
	}
	return content, file.SHA, nil
}

func (client *Client) AddIssueComment(ctx context.Context, number int, body string) error {
	_, err := client.request(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/issues/%d/comments", client.Owner, client.Repository, number), map[string]string{"body": body}, nil)
	return err
}

func (client *Client) SetWorkflowLabel(ctx context.Context, number int, label string) error {
	if _, ok := workflowLabels[label]; !ok {
		return fmt.Errorf("unknown VI workflow label %q", label)
	}
	if err := client.ensureLabel(ctx, label); err != nil {
		return err
	}
	issue, err := client.GetIssue(ctx, number)
	if err != nil {
		return err
	}
	labels := make([]string, 0, len(issue.Labels)+1)
	for _, existing := range issue.Labels {
		if _, isWorkflowLabel := workflowLabels[existing.Name]; !isWorkflowLabel {
			labels = append(labels, existing.Name)
		}
	}
	labels = append(labels, label)
	_, err = client.request(ctx, http.MethodPut, fmt.Sprintf("/repos/%s/%s/issues/%d/labels", client.Owner, client.Repository, number), map[string][]string{"labels": labels}, nil)
	return err
}

func (client *Client) ensureLabel(ctx context.Context, name string) error {
	path := fmt.Sprintf("/repos/%s/%s/labels/%s", client.Owner, client.Repository, url.PathEscape(name))
	status, err := client.request(ctx, http.MethodGet, path, nil, nil)
	if err == nil {
		return nil
	}
	if status != http.StatusNotFound {
		return err
	}
	_, err = client.request(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/labels", client.Owner, client.Repository), map[string]string{
		"name":        name,
		"color":       "1d76db",
		"description": "VI documentation workflow state",
	}, nil)
	return err
}

func (client *Client) ReopenIssue(ctx context.Context, number int) error {
	_, err := client.request(ctx, http.MethodPatch, fmt.Sprintf("/repos/%s/%s/issues/%d", client.Owner, client.Repository, number), map[string]string{"state": "open"}, nil)
	return err
}

func (client *Client) CreateDocumentationPR(ctx context.Context, issueNumber int, title, body string, files map[string]string) (CreatedPullRequest, error) {
	repositoryPath := fmt.Sprintf("/repos/%s/%s", client.Owner, client.Repository)
	var repository Repository
	if _, err := client.request(ctx, http.MethodGet, repositoryPath, nil, &repository); err != nil {
		return CreatedPullRequest{}, err
	}
	branch := fmt.Sprintf("docs/vi-%d", issueNumber)
	var existing []CreatedPullRequest
	query := url.Values{"state": []string{"open"}, "head": []string{client.Owner + ":" + branch}, "base": []string{repository.DefaultBranch}}
	_, listErr := client.request(ctx, http.MethodGet, repositoryPath+"/pulls?"+query.Encode(), nil, &existing)
	if listErr != nil {
		return CreatedPullRequest{}, listErr
	}
	if len(existing) > 0 {
		return existing[0], nil
	}

	var baseRef struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	refPath := repositoryPath + "/git/ref/heads/" + branch
	status, err := client.request(ctx, http.MethodGet, refPath, nil, &baseRef)
	if err != nil && status != http.StatusNotFound {
		return CreatedPullRequest{}, err
	}
	if status == http.StatusNotFound {
		var baseBranch struct {
			Commit struct {
				SHA string `json:"sha"`
			} `json:"commit"`
		}
		if _, err := client.request(ctx, http.MethodGet, repositoryPath+"/branches/"+repository.DefaultBranch, nil, &baseBranch); err != nil {
			return CreatedPullRequest{}, err
		}
		_, err := client.request(ctx, http.MethodPost, repositoryPath+"/git/refs", map[string]string{"ref": "refs/heads/" + branch, "sha": baseBranch.Commit.SHA}, nil)
		if err != nil {
			return CreatedPullRequest{}, err
		}
		baseRef.Object.SHA = baseBranch.Commit.SHA
	}

	for path, content := range files {
		_, fileSHA, err := client.GetFile(ctx, path, branch)
		if err != nil {
			return CreatedPullRequest{}, err
		}
		payload := map[string]string{
			"message": fmt.Sprintf("docs: update %s for VI #%d", path, issueNumber),
			"content": base64.StdEncoding.EncodeToString([]byte(content)),
			"branch":  branch,
			"sha":     fileSHA,
		}
		_, err = client.request(ctx, http.MethodPut, repositoryPath+"/contents/"+path, payload, nil)
		if err != nil {
			return CreatedPullRequest{}, err
		}
	}

	payload := map[string]string{"title": title, "body": body, "head": branch, "base": repository.DefaultBranch}
	var pull CreatedPullRequest
	_, err = client.request(ctx, http.MethodPost, repositoryPath+"/pulls", payload, &pull)
	if err != nil {
		return CreatedPullRequest{}, err
	}
	return pull, nil
}
