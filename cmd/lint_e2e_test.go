//go:build e2e

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
package cmd_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var runnBin string

func TestMain(m *testing.M) {
	// Build runn binary before running tests
	runnBin = filepath.Join(os.TempDir(), "runn-test")
	cmd := exec.Command("go", "build", "-o", runnBin, "./runn")
	cmd.Dir = filepath.Join("..", "cmd")
	if err := cmd.Run(); err != nil {
		panic("failed to build runn: " + err.Error())
	}
	code := m.Run()
	os.Remove(runnBin)
	os.Exit(code)
}

func TestLintE2E_ValidFiles(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{"valid runbook", "testdata/valid.yml"},
		{"formatted runbook", "testdata/formatted.yml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(runnBin, "lint", tt.file)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("lint should pass for %s, but got error: %v\noutput: %s", tt.file, err, output)
			}
		})
	}
}

func TestLintE2E_InvalidFiles(t *testing.T) {
	tests := []struct {
		name        string
		file        string
		wantContain string
	}{
		{
			name:        "invalid indent",
			file:        "testdata/invalid_indent.yml",
			wantContain: "value is not allowed",
		},
		{
			name:        "invalid colon",
			file:        "testdata/invalid_colon.yml",
			wantContain: "unexpected key name",
		},
		{
			name:        "invalid quote",
			file:        "testdata/invalid_quote.yml",
			wantContain: "could not find end character",
		},
		{
			name:        "invalid yaml syntax",
			file:        "testdata/invalid_yaml.yml",
			wantContain: "non-map value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(runnBin, "lint", tt.file)
			output, err := cmd.CombinedOutput()
			if err == nil {
				t.Errorf("lint should fail for %s, but succeeded", tt.file)
			}
			if !strings.Contains(string(output), tt.wantContain) {
				t.Errorf("output should contain %q, but got: %s", tt.wantContain, output)
			}
		})
	}
}

func TestLintE2E_NonExistentFile(t *testing.T) {
	cmd := exec.Command(runnBin, "lint", "testdata/nonexistent.yml")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("lint should fail for non-existent file")
	}
	if !strings.Contains(string(output), "no such file") {
		t.Errorf("output should contain 'no such file', but got: %s", output)
	}
}

func TestLintE2E_MultipleFiles(t *testing.T) {
	// All valid
	cmd := exec.Command(runnBin, "lint", "testdata/valid.yml", "testdata/formatted.yml")
	if err := cmd.Run(); err != nil {
		t.Errorf("lint should pass for multiple valid files: %v", err)
	}

	// Mix of valid and invalid
	cmd = exec.Command(runnBin, "lint", "testdata/valid.yml", "testdata/invalid_indent.yml")
	if err := cmd.Run(); err == nil {
		t.Error("lint should fail when any file is invalid")
	}
}
