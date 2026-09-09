package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/google/go-github/v89/github"
)

func TestEnvironmentProtectionFallback(t *testing.T) {
	t.Parallel()

	reviewers := []*github.EnvReviewers{{Type: new("User"), ID: new(int64(123))}}
	tests := []struct {
		name          string
		initialStatus int
		retryStatus   int
		input         github.CreateUpdateEnvironment
		wantRequests  int
		wantError     bool
	}{
		{
			name: "Team fallback preserves enabled admin bypass", initialStatus: http.StatusUnprocessableEntity,
			retryStatus: http.StatusOK, input: github.CreateUpdateEnvironment{CanAdminsBypass: new(true)}, wantRequests: 2,
		},
		{
			name: "fallback does not discard disabled admin bypass", initialStatus: http.StatusUnprocessableEntity,
			retryStatus: http.StatusOK, input: github.CreateUpdateEnvironment{CanAdminsBypass: new(false)}, wantRequests: 2,
		},
		{
			name: "successful request needs no fallback", initialStatus: http.StatusOK,
			input: github.CreateUpdateEnvironment{CanAdminsBypass: new(true)}, wantRequests: 1,
		},
		{
			name: "authorization failure is not retried", initialStatus: http.StatusForbidden,
			input: github.CreateUpdateEnvironment{CanAdminsBypass: new(true)}, wantRequests: 1, wantError: true,
		},
		{
			name: "required reviewers are not silently removed", initialStatus: http.StatusUnprocessableEntity,
			input: github.CreateUpdateEnvironment{Reviewers: reviewers}, wantRequests: 1, wantError: true,
		},
		{
			name: "wait timer is not silently removed", initialStatus: http.StatusUnprocessableEntity,
			input: github.CreateUpdateEnvironment{WaitTimer: new(1)}, wantRequests: 1, wantError: true,
		},
		{
			name: "self review prevention is not silently removed", initialStatus: http.StatusUnprocessableEntity,
			input: github.CreateUpdateEnvironment{PreventSelfReview: new(true)}, wantRequests: 1, wantError: true,
		},
		{
			name: "fallback rejection is returned", initialStatus: http.StatusUnprocessableEntity,
			retryStatus: http.StatusForbidden, input: github.CreateUpdateEnvironment{CanAdminsBypass: new(true)},
			wantRequests: 2, wantError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			branchPolicy := &github.BranchPolicy{ProtectedBranches: new(false), CustomBranchPolicies: new(true)}
			test.input.DeploymentBranchPolicy = branchPolicy
			var requests atomic.Int32
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount := requests.Add(1)
				if r.Method != http.MethodPut || r.URL.EscapedPath() != "/repos/owner/repository/environments/production%2Fblue" {
					t.Errorf("unexpected environment request: %s %s", r.Method, r.URL.EscapedPath())
				}
				w.Header().Set("Content-Type", "application/json")
				if requestCount == 1 {
					w.WriteHeader(test.initialStatus)
				} else if test.retryStatus == 0 {
					t.Errorf("unexpected retry after status %d", test.initialStatus)
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					var actual map[string]any
					if err := json.NewDecoder(r.Body).Decode(&actual); err != nil {
						t.Errorf("decode fallback body: %v", err)
					}
					expected := map[string]any{
						"can_admins_bypass": test.input.GetCanAdminsBypass(),
						"deployment_branch_policy": map[string]any{
							"protected_branches": false, "custom_branch_policies": true,
						},
					}
					if !reflect.DeepEqual(actual, expected) {
						t.Errorf("fallback payload mismatch: got %#v, want %#v", actual, expected)
					}
					w.WriteHeader(test.retryStatus)
				}
				if _, err := fmt.Fprint(w, `{}`); err != nil {
					t.Errorf("write test API response: %v", err)
				}
			})
			server := httptest.NewServer(handler)
			t.Cleanup(server.Close)
			baseURL := server.URL + "/"
			client, err := github.NewClient(github.WithHTTPClient(server.Client()), github.WithURLs(&baseURL, nil))
			if err != nil {
				t.Fatalf("configure test GitHub client: %v", err)
			}

			err = applyRepositoryEnvironment(context.Background(), client, "owner", "repository", "production/blue", &test.input)

			if (err != nil) != test.wantError {
				t.Errorf("apply environment error: got %v, want error=%v", err, test.wantError)
			}
			if actual := int(requests.Load()); actual != test.wantRequests {
				t.Errorf("environment request count: got %d, want %d", actual, test.wantRequests)
			}
		})
	}
}
