package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func checkGHAvailable() error {
	_, err := exec.LookPath("gh")
	if err != nil {
		return fmt.Errorf("gh CLI not found in PATH: %w", err)
	}
	return nil
}

func canViewRepo(repo string) bool {
	cmd := exec.Command("gh", "repo", "view", repo)
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func canAccessUserEmails() bool {
	cmd := exec.Command("gh", "api", "user/emails")
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func runGHAuthRefresh() error {
	cmd := exec.Command("gh", "auth", "refresh", "-h", "github.com", "-s", "user")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func cloneRepo(repo string) (string, error) {
	// Determine the directory name that gh repo clone will create
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid repo format: %s", repo)
	}
	repoName := parts[1]

	cmd := exec.Command("gh", "repo", "clone", repo)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to clone repo: %w", err)
	}

	// gh repo clone creates a directory named after the repo in the current directory
	cloneDir, err := filepath.Abs(repoName)
	if err != nil {
		return "", fmt.Errorf("failed to determine clone directory: %w", err)
	}
	return cloneDir, nil
}

// User represents a GitHub user API response.
type User struct {
	Login string `json:"login"`
}

func getGHUser() (string, error) {
	out, err := exec.Command("gh", "api", "user").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get GitHub user: %w", err)
	}
	var u User
	if err := json.Unmarshal(out, &u); err != nil {
		return "", fmt.Errorf("failed to parse user response: %w", err)
	}
	if u.Login == "" {
		return "", fmt.Errorf("GitHub user login not found in response")
	}
	return u.Login, nil
}

// Email represents a GitHub user email entry.
type Email struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func getGHEmail() (string, error) {
	out, err := exec.Command("gh", "api", "user/emails").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get GitHub emails: %w", err)
	}

	var emails []Email
	if err := json.Unmarshal(out, &emails); err != nil {
		return "", fmt.Errorf("failed to parse emails response: %w", err)
	}

	// Prefer github.com noreply emails, then fall back to the first available email.
	githubPattern := regexp.MustCompile(`github\.com`)
	var fallback string
	for _, e := range emails {
		if githubPattern.MatchString(e.Email) {
			return e.Email, nil
		}
		if fallback == "" {
			fallback = e.Email
		}
	}

	if fallback != "" {
		return fallback, nil
	}
	return "", fmt.Errorf("no suitable email address found")
}

func setGitConfig(dir, name, email string) error {
	cmds := [][]string{
		{"git", "config", "--local", "user.name", name},
		{"git", "config", "--local", "user.email", email},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run %v: %w", args, err)
		}
	}
	return nil
}

func run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: ghclone <owner/repo>")
	}
	repo := args[0]
	if !strings.Contains(repo, "/") {
		return fmt.Errorf("invalid repo format %q, expected owner/repo", repo)
	}

	if err := checkGHAvailable(); err != nil {
		return err
	}

	// Check if authentication is sufficient
	if !canViewRepo(repo) || !canAccessUserEmails() {
		fmt.Println("Insufficient GitHub permissions. Refreshing authentication...")
		if err := runGHAuthRefresh(); err != nil {
			return fmt.Errorf("gh auth refresh failed: %w", err)
		}
	}

	cloneDir, err := cloneRepo(repo)
	if err != nil {
		return err
	}

	ghUser, err := getGHUser()
	if err != nil {
		return err
	}

	ghEmail, err := getGHEmail()
	if err != nil {
		return err
	}

	if err := setGitConfig(cloneDir, ghUser, ghEmail); err != nil {
		return err
	}

	fmt.Printf("Cloned %s and configured git user.name=%q user.email=%q\n", repo, ghUser, ghEmail)
	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
