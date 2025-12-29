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
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFmtE2E_Output(t *testing.T) {
	cmd := exec.Command(runnBin, "fmt", "testdata/unformatted.yml")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("fmt should pass: %v", err)
	}

	// Check key order: desc should come before runners, runners before steps
	descIdx := bytes.Index(output, []byte("desc:"))
	runnersIdx := bytes.Index(output, []byte("runners:"))
	stepsIdx := bytes.Index(output, []byte("steps:"))

	if descIdx == -1 || runnersIdx == -1 || stepsIdx == -1 {
		t.Fatalf("output missing expected keys: %s", output)
	}

	if descIdx > runnersIdx {
		t.Error("desc should come before runners")
	}
	if runnersIdx > stepsIdx {
		t.Error("runners should come before steps")
	}
}

func TestFmtE2E_WriteOption(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yml")

	content, err := os.ReadFile("testdata/unformatted.yml")
	if err != nil {
		t.Fatalf("failed to read testdata: %v", err)
	}
	if err := os.WriteFile(tmpFile, content, 0o644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// Run fmt with --write
	cmd := exec.Command(runnBin, "fmt", "--write", tmpFile)
	if err := cmd.Run(); err != nil {
		t.Fatalf("fmt --write should pass: %v", err)
	}

	// Read formatted file
	formatted, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read formatted file: %v", err)
	}

	// Check key order
	descIdx := bytes.Index(formatted, []byte("desc:"))
	runnersIdx := bytes.Index(formatted, []byte("runners:"))
	stepsIdx := bytes.Index(formatted, []byte("steps:"))

	if descIdx > runnersIdx {
		t.Error("desc should come before runners in written file")
	}
	if runnersIdx > stepsIdx {
		t.Error("runners should come before steps in written file")
	}
}

func TestFmtE2E_InvalidFile(t *testing.T) {
	cmd := exec.Command(runnBin, "fmt", "testdata/invalid_yaml.yml")
	if err := cmd.Run(); err == nil {
		t.Error("fmt should fail for invalid YAML")
	}
}

func TestFmtE2E_NonExistentFile(t *testing.T) {
	cmd := exec.Command(runnBin, "fmt", "testdata/nonexistent.yml")
	if err := cmd.Run(); err == nil {
		t.Error("fmt should fail for non-existent file")
	}
}
