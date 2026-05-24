package gitlab

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"git-config-manager/internal/config"
	"git-config-manager/internal/provider"
	"git-config-manager/pkg/logger"
)

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return NewClient(config.ProviderConfig{APIURL: server.URL}, logger.New(logger.LevelError, os.Stderr))
}

func TestGetUser_UsesPrivateTokenForPAT(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("PRIVATE-TOKEN"); got != "glpat" {
			t.Fatalf("PRIVATE-TOKEN = %q", got)
		}
		json.NewEncoder(w).Encode(User{Username: "jane", Name: "Jane Doe"})
	})
	client.SetToken("glpat")

	user, err := client.GetUser(context.Background())
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if user.Username != "jane" {
		t.Fatalf("Username = %q", user.Username)
	}
}

func TestGetUser_UsesBearerForOAuthToken(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer oauth-token" {
			t.Fatalf("Authorization = %q", got)
		}
		json.NewEncoder(w).Encode(User{Username: "oauth-user"})
	})
	client.SetTokenSet(provider.TokenSet{AccessToken: "oauth-token", AuthMethod: provider.AuthMethodOAuthDevice, TokenType: "bearer"})

	if _, err := client.GetUser(context.Background()); err != nil {
		t.Fatalf("GetUser: %v", err)
	}
}

func TestUploadSSHKey(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/user/keys" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload["title"] != "work" || payload["key"] == "" {
			t.Fatalf("payload = %#v", payload)
		}
		w.WriteHeader(http.StatusCreated)
	})
	client.SetToken("glpat")

	if err := client.UploadSSHKey(context.Background(), "work", "ssh-ed25519 AAAA"); err != nil {
		t.Fatalf("UploadSSHKey: %v", err)
	}
}

func TestSSHKeyExists_NormalizesKeyComments(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]SSHKeyResponse{{ID: 1, Key: "ssh-ed25519 AAAA old-comment"}})
	})
	client.SetToken("glpat")

	exists, err := client.SSHKeyExists(context.Background(), "ssh-ed25519 AAAA new-comment")
	if err != nil {
		t.Fatalf("SSHKeyExists: %v", err)
	}
	if !exists {
		t.Fatal("expected key to exist")
	}
}

func TestDeleteSSHKey(t *testing.T) {
	deleteCalled := false
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/user/keys":
			json.NewEncoder(w).Encode([]SSHKeyResponse{{ID: 7, Key: "ssh-ed25519 AAAA old-comment"}})
		case r.Method == http.MethodDelete && r.URL.Path == "/user/keys/7":
			deleteCalled = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})
	client.SetToken("glpat")

	deleted, err := client.DeleteSSHKey(context.Background(), "ssh-ed25519 AAAA new-comment")
	if err != nil {
		t.Fatalf("DeleteSSHKey: %v", err)
	}
	if !deleted || !deleteCalled {
		t.Fatal("expected SSH key to be deleted")
	}
}

func TestUploadGPGKey(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/user/gpg_keys" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload["key"] == "" {
			t.Fatalf("payload = %#v", payload)
		}
		w.WriteHeader(http.StatusCreated)
	})
	client.SetToken("glpat")

	if err := client.UploadGPGKey(context.Background(), "-----BEGIN PGP PUBLIC KEY BLOCK-----"); err != nil {
		t.Fatalf("UploadGPGKey: %v", err)
	}
}

func TestDeleteGPGKey(t *testing.T) {
	deleteCalled := false
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/user/gpg_keys":
			json.NewEncoder(w).Encode([]GPGKeyResponse{{ID: 9, KeyID: "ABC123", Fingerprint: "FINGERPRINT"}})
		case r.Method == http.MethodDelete && r.URL.Path == "/user/gpg_keys/9":
			deleteCalled = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})
	client.SetToken("glpat")

	deleted, err := client.DeleteGPGKey(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("DeleteGPGKey: %v", err)
	}
	if !deleted || !deleteCalled {
		t.Fatal("expected GPG key to be deleted")
	}
}

func TestApiError(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"forbidden"}`))
	})
	client.SetToken("bad")

	if _, err := client.GetUser(context.Background()); err == nil {
		t.Fatal("expected API error")
	}
}
