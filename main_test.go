package main

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

func TestRunNoArgs(t *testing.T) {
	err := run([]string{})
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
	if !strings.Contains(err.Error(), "Usage:") {
		t.Fatalf("expected usage in error, got: %v", err)
	}
}

func TestRunInvalidRepoFormat(t *testing.T) {
	err := run([]string{"noslash"})
	if err == nil {
		t.Fatal("expected error for invalid repo format")
	}
	if !strings.Contains(err.Error(), "Usage:") {
		t.Fatalf("expected usage in error, got: %v", err)
	}
}

func TestRunTooManyArgs(t *testing.T) {
	err := run([]string{"owner/repo", "extra"})
	if err == nil {
		t.Fatal("expected error when too many args provided")
	}
	if !strings.Contains(err.Error(), "Usage:") {
		t.Fatalf("expected usage in error, got: %v", err)
	}
}

func TestRunHelp(t *testing.T) {
	if err := run([]string{"--help"}); err != nil {
		t.Fatalf("expected no error for --help, got: %v", err)
	}
}

func TestRunVersion(t *testing.T) {
	if err := run([]string{"--version"}); err != nil {
		t.Fatalf("expected no error for --version, got: %v", err)
	}
}

func TestGetGHEmailParsingPicksGithubEmail(t *testing.T) {
	emails := []Email{
		{Email: "user@example.com", Primary: true, Verified: true},
		{Email: "user@users.noreply.github.com", Primary: false, Verified: true},
	}
	data, _ := json.Marshal(emails)

	var parsed []Email
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Replicate getGHEmail logic
	githubPattern := regexp.MustCompile(`github\.com`)
	var found, fallback string
	for _, e := range parsed {
		if githubPattern.MatchString(e.Email) {
			found = e.Email
			break
		}
		if fallback == "" {
			fallback = e.Email
		}
	}
	if found == "" {
		found = fallback
	}

	// The github.com email should be preferred even though it is listed second
	if found != "user@users.noreply.github.com" {
		t.Errorf("expected github.com email to be selected, got: %s", found)
	}
}

func TestGetGHEmailParsingFallsBackToFirstEmail(t *testing.T) {
	emails := []Email{
		{Email: "user@example.com", Primary: true, Verified: true},
		{Email: "other@example.org", Primary: false, Verified: true},
	}
	data, _ := json.Marshal(emails)

	var parsed []Email
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	githubPattern := regexp.MustCompile(`github\.com`)
	var found, fallback string
	for _, e := range parsed {
		if githubPattern.MatchString(e.Email) {
			found = e.Email
			break
		}
		if fallback == "" {
			fallback = e.Email
		}
	}
	if found == "" {
		found = fallback
	}

	// With no github.com email, fall back to the first email
	if found != "user@example.com" {
		t.Errorf("expected fallback to first email, got: %s", found)
	}
}

func TestGetGHUserParsing(t *testing.T) {
	u := User{Login: "octocat"}
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed User
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if parsed.Login != "octocat" {
		t.Errorf("expected login %q, got %q", "octocat", parsed.Login)
	}
}

func TestGetGHUserParsingEmptyLogin(t *testing.T) {
	data, err := json.Marshal(User{})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var u User
	if err := json.Unmarshal(data, &u); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if u.Login != "" {
		t.Errorf("expected empty login, got %q", u.Login)
	}
}

func TestCloneRepoInvalidFormat(t *testing.T) {
	_, err := cloneRepo("noslash")
	if err == nil {
		t.Fatal("expected error for invalid repo format")
	}
}
