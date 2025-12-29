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
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	rpEndpoint    string
	rpToken       string
	rpProject     string
	rpLaunchName  string
	rpDescription string
)

// launchImportRq represents the request body for ReportPortal launch import.
type launchImportRq struct {
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Attributes  []attribute `json:"attributes,omitempty"`
}

type attribute struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value"`
}

// reportportalCmd represents the reportportal command.
var reportportalCmd = &cobra.Command{
	Use:   "reportportal [JUNIT_XML_FILE]",
	Short: "send test results to ReportPortal",
	Long: `Send JUnit XML test results to ReportPortal.

Environment variables:
  RP_ENDPOINT  ReportPortal URL (e.g., http://localhost:8080)
  RP_TOKEN     ReportPortal API token
  RP_PROJECT   Project name (e.g., superadmin_personal)

Example:
  runn reportportal --endpoint http://localhost:8080 --token xxx --project my_project report.xml

  # Or using environment variables
  export RP_ENDPOINT=http://localhost:8080
  export RP_TOKEN=xxx
  export RP_PROJECT=my_project
  runn reportportal report.xml`,
	Aliases: []string{"rp"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		xmlFile := args[0]

		// Get configuration from flags or environment variables
		endpoint := rpEndpoint
		if endpoint == "" {
			endpoint = os.Getenv("RP_ENDPOINT")
		}
		if endpoint == "" {
			return errors.New("endpoint is required: use --endpoint flag or RP_ENDPOINT environment variable")
		}

		token := rpToken
		if token == "" {
			token = os.Getenv("RP_TOKEN")
		}
		if token == "" {
			return errors.New("token is required: use --token flag or RP_TOKEN environment variable")
		}

		project := rpProject
		if project == "" {
			project = os.Getenv("RP_PROJECT")
		}
		if project == "" {
			return errors.New("project is required: use --project flag or RP_PROJECT environment variable")
		}

		// Check if file exists
		if _, err := os.Stat(xmlFile); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", xmlFile)
		}

		// Set default launch name
		launchName := rpLaunchName
		if launchName == "" {
			launchName = strings.TrimSuffix(filepath.Base(xmlFile), filepath.Ext(xmlFile))
		}

		// Send to ReportPortal
		return sendToReportPortal(cmd, endpoint, token, project, xmlFile, launchName, rpDescription)
	},
}

func sendToReportPortal(cmd *cobra.Command, endpoint, token, project, xmlFile, launchName, description string) error {
	// Open the XML file
	file, err := os.Open(xmlFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file
	part, err := writer.CreateFormFile("file", filepath.Base(xmlFile))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Add launchImportRq
	launchRq := launchImportRq{
		Name:        launchName,
		Description: description,
		Attributes: []attribute{
			{Key: "source", Value: "runn"},
			{Key: "type", Value: "junit-import"},
		},
	}
	launchRqJSON, err := json.Marshal(launchRq)
	if err != nil {
		return fmt.Errorf("failed to marshal launch request: %w", err)
	}

	if err := writer.WriteField("launchImportRq", string(launchRqJSON)); err != nil {
		return fmt.Errorf("failed to write launch request: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	// Build URL
	url := fmt.Sprintf("%s/api/v1/plugin/%s/junit/import", strings.TrimSuffix(endpoint, "/"), project)

	// Create request
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	fmt.Fprintf(cmd.OutOrStdout(), "Sending test results to ReportPortal...\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  Endpoint: %s\n", endpoint)
	fmt.Fprintf(cmd.OutOrStdout(), "  Project:  %s\n", project)
	fmt.Fprintf(cmd.OutOrStdout(), "  File:     %s\n", xmlFile)
	fmt.Fprintf(cmd.OutOrStdout(), "  Launch:   %s\n", launchName)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Fprintf(cmd.OutOrStdout(), "Success! Response: %s\n", string(body))
		return nil
	}

	return fmt.Errorf("failed to upload: HTTP %d: %s", resp.StatusCode, string(body))
}

func init() {
	rootCmd.AddCommand(reportportalCmd)
	reportportalCmd.Flags().StringVar(&rpEndpoint, "endpoint", "", "ReportPortal endpoint URL (or RP_ENDPOINT env)")
	reportportalCmd.Flags().StringVar(&rpToken, "token", "", "ReportPortal API token (or RP_TOKEN env)")
	reportportalCmd.Flags().StringVar(&rpProject, "project", "", "ReportPortal project name (or RP_PROJECT env)")
	reportportalCmd.Flags().StringVar(&rpLaunchName, "name", "", "Launch name (default: filename)")
	reportportalCmd.Flags().StringVar(&rpDescription, "description", "", "Launch description")
}
