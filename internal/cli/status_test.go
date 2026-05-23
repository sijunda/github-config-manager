package cli

import (
	"testing"
)

func TestPadRight(t *testing.T) {
	if got := padRight("hi", 5); got != "hi   " {
		t.Fatalf("padRight(hi, 5) = %q", got)
	}
	if got := padRight("hello", 3); got != "hello" {
		t.Fatalf("padRight(hello, 3) = %q", got)
	}
}

func TestStatusCmd_NoProfiles(t *testing.T) {
	env := setupTestEnv(t)
	_ = env
	if err := executeCmd("status"); err != nil {
		t.Fatalf("status: %v", err)
	}
}

func TestStatusCmd_WithProfiles(t *testing.T) {
	env := setupTestEnv(t)
	env.createProfile("work", "work@test.com")
	env.createProfile("personal", "personal@test.com")
	env.fakeGitScript("exit 0")
	_ = executeCmd("use", "work", "--global")

	if err := executeCmd("status"); err != nil {
		t.Fatalf("status with profiles: %v", err)
	}
}

func TestStatusCmd_WithSSHKey(t *testing.T) {
	env := setupTestEnv(t)
	env.createProfileWithSSH("work", "work@test.com")
	env.fakeGitScript("exit 0")
	_ = executeCmd("use", "work", "--global")

	if err := executeCmd("status"); err != nil {
		t.Fatalf("status with ssh: %v", err)
	}
}

func TestStatusCmd_WithToken(t *testing.T) {
	env := setupTestEnv(t)
	env.createProfile("work", "work@test.com")
	_ = ctr.GitHubClient.SaveToken("work", "ghp_faketoken")
	env.fakeGitScript("exit 0")
	_ = executeCmd("use", "work", "--global")

	if err := executeCmd("status"); err != nil {
		t.Fatalf("status with token: %v", err)
	}
}

func TestQuickVerifyToken_Invalid(t *testing.T) {
	err := quickVerifyToken("invalid_token", "https://api.github.com")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
