package runn

import (
	"fmt"
	"io"
	"net/http"

	"github.com/fatih/color"
)

var _ Capturer = (*cmdOut)(nil)

type cmdOut struct {
	out     io.Writer
	verbose bool
	errs    error
	green   func(a ...interface{}) string
	yellow  func(a ...interface{}) string
	red     func(a ...interface{}) string
}

func NewCmdOut(out io.Writer, verbose bool) *cmdOut {
	return &cmdOut{
		out:     out,
		verbose: verbose,
		green:   color.New(color.FgGreen).SprintFunc(),
		yellow:  color.New(color.FgYellow).SprintFunc(),
		red:     color.New(color.FgRed).SprintFunc(),
	}
}

func (d *cmdOut) CaptureStart(ids IDs, bookPath, desc string) {}
func (d *cmdOut) CaptureResult(ids IDs, result *RunResult) {
	switch {
	case result.Err != nil:
		_, _ = fmt.Fprintf(d.out, "%s (%s) ... %v\n", result.Desc, ShortenPath(result.Path), d.red(result.Err))
	case result.Skipped:
		_, _ = fmt.Fprintf(d.out, "%s (%s) ... %s\n", result.Desc, ShortenPath(result.Path), d.yellow("skip"))
	default:
		_, _ = fmt.Fprintf(d.out, "%s (%s) ... %s\n", result.Desc, ShortenPath(result.Path), d.green("ok"))
	}
	if d.verbose {
		for _, sr := range result.StepResults {
			switch {
			case sr.Err != nil:
				_, _ = fmt.Fprintf(d.out, "  %s (%s) ... %v\n", sr.Desc, sr.Key, d.red(sr.Err))
			case sr.Skipped:
				_, _ = fmt.Fprintf(d.out, "  %s (%s) ... %s\n", sr.Desc, sr.Key, d.yellow("skip"))
			default:
				_, _ = fmt.Fprintf(d.out, "  %s (%s) ... %s\n", sr.Desc, sr.Key, d.green("ok"))
			}
		}
	}
}
func (d *cmdOut) CaptureEnd(ids IDs, bookPath, desc string) {}

func (d *cmdOut) CaptureHTTPRequest(name string, req *http.Request)                  {}
func (d *cmdOut) CaptureHTTPResponse(name string, res *http.Response)                {}
func (d *cmdOut) CaptureGRPCStart(name string, typ GRPCType, service, method string) {}
func (d *cmdOut) CaptureGRPCRequestHeaders(h map[string][]string)                    {}
func (d *cmdOut) CaptureGRPCRequestMessage(m map[string]interface{})                 {}
func (d *cmdOut) CaptureGRPCResponseStatus(status int)                               {}
func (d *cmdOut) CaptureGRPCResponseHeaders(h map[string][]string)                   {}
func (d *cmdOut) CaptureGRPCResponseMessage(m map[string]interface{})                {}
func (d *cmdOut) CaptureGRPCResponseTrailers(t map[string][]string)                  {}
func (d *cmdOut) CaptureGRPCClientClose()                                            {}
func (d *cmdOut) CaptureGRPCEnd(name string, typ GRPCType, service, method string)   {}
func (d *cmdOut) CaptureCDPStart(name string)                                        {}
func (d *cmdOut) CaptureCDPAction(a CDPAction)                                       {}
func (d *cmdOut) CaptureCDPResponse(a CDPAction, res map[string]interface{})         {}
func (d *cmdOut) CaptureCDPEnd(name string)                                          {}
func (d *cmdOut) CaptureSSHCommand(command string)                                   {}
func (d *cmdOut) CaptureSSHStdout(stdout string)                                     {}
func (d *cmdOut) CaptureSSHStderr(stderr string)                                     {}
func (d *cmdOut) CaptureDBStatement(name string, stmt string)                        {}
func (d *cmdOut) CaptureDBResponse(name string, res *DBResponse)                     {}
func (d *cmdOut) CaptureExecCommand(command string)                                  {}
func (d *cmdOut) CaptureExecStdin(stdin string)                                      {}
func (d *cmdOut) CaptureExecStdout(stdout string)                                    {}
func (d *cmdOut) CaptureExecStderr(stderr string)                                    {}
func (d *cmdOut) SetCurrentIDs(ids IDs)                                              {}
func (d *cmdOut) Errs() error {
	return d.errs
}
