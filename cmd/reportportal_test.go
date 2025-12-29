/*
Copyright © 2022 Ken'ichiro Oyama <k1lowxb@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReportPortalCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no arguments",
			args:    []string{"reportportal"},
			wantErr: true,
		},
		{
			name:    "non-existent file",
			args:    []string{"reportportal", "testdata/nonexistent.xml"},
			wantErr: true,
		},
		{
			name:    "missing endpoint",
			args:    []string{"reportportal", "testdata/valid.yml"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数をクリア
			t.Setenv("RP_ENDPOINT", "")
			t.Setenv("RP_TOKEN", "")
			t.Setenv("RP_PROJECT", "")

			rootCmd.SetArgs(tt.args)
			rootCmd.SetOut(&bytes.Buffer{})
			rootCmd.SetErr(&bytes.Buffer{})

			err := rootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("reportportal command error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReportPortalCommandWithFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name: "with all required flags but non-existent file",
			args: []string{
				"reportportal",
				"--endpoint", "http://localhost:8080",
				"--token", "test-token",
				"--project", "test-project",
				"testdata/nonexistent.xml",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd.SetArgs(tt.args)
			rootCmd.SetOut(&bytes.Buffer{})
			rootCmd.SetErr(&bytes.Buffer{})

			err := rootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("reportportal command error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReportPortalCommandWithMockServer(t *testing.T) {
	// Create a mock ReportPortal server
	var receivedAuth string
	var receivedContentType string
	var receivedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		receivedContentType = r.Header.Get("Content-Type")

		body, _ := io.ReadAll(r.Body)
		receivedBody = body

		// Check endpoint path
		if !strings.HasSuffix(r.URL.Path, "/junit/import") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Launch with id = 123 is successfully imported",
		})
	}))
	defer server.Close()

	// Create a test XML file (use existing valid.yml as it's just for file existence check)
	testFile := "testdata/valid.yml"

	out := &bytes.Buffer{}
	rootCmd.SetArgs([]string{
		"reportportal",
		"--endpoint", server.URL,
		"--token", "test-token-123",
		"--project", "test-project",
		"--name", "Test Launch",
		"--description", "Test Description",
		testFile,
	})
	rootCmd.SetOut(out)
	rootCmd.SetErr(&bytes.Buffer{})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify authorization header
	if receivedAuth != "Bearer test-token-123" {
		t.Errorf("expected auth 'Bearer test-token-123', got '%s'", receivedAuth)
	}

	// Verify content type is multipart
	if !strings.HasPrefix(receivedContentType, "multipart/form-data") {
		t.Errorf("expected multipart/form-data content type, got '%s'", receivedContentType)
	}

	// Verify body contains file data
	if len(receivedBody) == 0 {
		t.Error("expected non-empty body")
	}

	// Verify output contains success message
	output := out.String()
	if !strings.Contains(output, "Success!") {
		t.Errorf("expected success message in output, got: %s", output)
	}
}

func TestReportPortalCommandWithMockServerError(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	testFile := "testdata/valid.yml"

	rootCmd.SetArgs([]string{
		"reportportal",
		"--endpoint", server.URL,
		"--token", "invalid-token",
		"--project", "test-project",
		testFile,
	})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}

	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401 error, got: %v", err)
	}
}

func TestReportPortalCommandWithEnvVars(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "success"})
	}))
	defer server.Close()

	// Reset flags
	rpEndpoint = ""
	rpToken = ""
	rpProject = ""
	rpLaunchName = ""
	rpDescription = ""

	// Set environment variables
	t.Setenv("RP_ENDPOINT", server.URL)
	t.Setenv("RP_TOKEN", "env-token")
	t.Setenv("RP_PROJECT", "env-project")

	testFile := "testdata/valid.yml"

	rootCmd.SetArgs([]string{"reportportal", testFile})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
