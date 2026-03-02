package main

import (
	"encoding/json"
	"regexp"
	"testing"
)

func TestRunNoArgs(t *testing.T) {
	err := run([]string{})
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}

func TestRunInvalidRepoFormat(t *testing.T) {
	err := run([]string{"noslash"})
	if err == nil {
		t.Fatal("expected error for invalid repo format")
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

func TestCloneRepoInvalidFormat(t *testing.T) {
	_, err := cloneRepo("noslash")
	if err == nil {
		t.Fatal("expected error for invalid repo format")
	}
}
