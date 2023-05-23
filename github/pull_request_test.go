package github_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	gogithub "github.com/google/go-github/v52/github"
	"github.com/ibihim/github-plays/github"
)

const (
	comment = "Hello, world!"
)

func setup() (*gogithub.Client, *httptest.Server) {
	mux := http.NewServeMux()

	// Mock the ref returnal for the PR.
	mux.HandleFunc("/repos/owner/repo/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"head":{"sha":"abc123"}}`)
	})

	// Mock the status response for the checks on the PR.
	mux.HandleFunc("/repos/owner/repo/commits/abc123/statuses", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"context":"ci: build","state":"success"},{"context":"ci: test","state":"pending"},{"context":"ci: deploy","state":"failure"}]`)
	})

	// Mock the response of the Issues.CreateComment method
	mux.HandleFunc("/repos/owner/repo/issues/1/comments", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Body string `json:"body"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if body.Body != comment {
			http.Error(w, "comment is not the expected comment", http.StatusBadRequest)
			return
		}

		fmt.Fprint(w, `{"body":"Hello, world!"}`)
	})

	server := httptest.NewServer(mux)

	client := gogithub.NewClient(nil)
	u, err := url.Parse(server.URL + "/")
	if err != nil {
		panic(err)
	}
	client.BaseURL = u

	return client, server
}

func teardown(server *httptest.Server) {
	server.Close()
}

func TestGetChecks(t *testing.T) {
	client, server := setup()
	t.Cleanup(func() {
		teardown(server)
	})

	prm := github.NewPullRequestManager(client, "owner", "repo", 1)

	success, pending, failure, err := prm.GetChecks(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(success, []string{"ci: build"}) {
		t.Errorf("expected success to be %v, but got %v", []string{"ci: build"}, success)
	}
	if !reflect.DeepEqual(pending, []string{"ci: test"}) {
		t.Errorf("expected pending to be %v, but got %v", []string{"ci: test"}, pending)
	}
	if !reflect.DeepEqual(failure, []string{"ci: deploy"}) {
		t.Errorf("expected failure to be %v, but got %v", []string{"ci: deploy"}, failure)
	}
}

func TestWriteComment(t *testing.T) {
	client, server := setup()
	defer teardown(server)

	prm := github.NewPullRequestManager(client, "owner", "repo", 1)
	if err := prm.WriteComment(context.Background(), comment); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
