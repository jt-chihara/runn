package runn

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/k1LoW/runn/testutil"
	"github.com/tenntenn/golden"
)

var testDebuggerHostRe = regexp.MustCompile(`(?s)Host:[^\r\n]+\r\n`)
var testDebuggerDateRe = regexp.MustCompile(`(?s)Date:[^\r\n]+\r\n`)

func TestDebugger(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/http.yml"},
		{"testdata/book/grpc.yml"},
		// {"testdata/book/cdp.yml"},
		{"testdata/book/pkg_go_dev.yml"},
		{"testdata/book/db.yml"},
		{"testdata/book/exec.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			if strings.Contains(tt.book, "cdp") && testutil.SkipCDPTest(t) {
				t.Skip("chrome not found")
			}
			out := new(bytes.Buffer)
			hs := testutil.HTTPServer(t)
			gs := testutil.GRPCServer(t, false)
			db, _ := testutil.SQLite(t)
			opts := []Option{
				Book(tt.book),
				HTTPRunner("req", hs.URL, hs.Client()),
				GrpcRunner("greq", gs.Conn()),
				DBRunner("db", db),
				Capture(NewDebugger(out)),
				Var("url", fmt.Sprintf("%s/form", hs.URL)),
			}
			o, err := New(opts...)
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}

			got := out.String()
			if strings.Contains(tt.book, "http.yml") {
				got = testDebuggerHostRe.ReplaceAllString(got, "Host: replace.example.com\r\n")
				got = testDebuggerDateRe.ReplaceAllString(got, "Date: Wed, 07 Sep 2022 06:28:20 GMT\r\n")
			}

			f := fmt.Sprintf("%s.debugger", filepath.Base(tt.book))
			if os.Getenv("UPDATE_GOLDEN") != "" {
				golden.Update(t, "testdata", f, got)
				return
			}

			if diff := golden.Diff(t, "testdata", f, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
