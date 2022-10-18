package runn

import (
	"errors"
	"fmt"
	"strings"

	"github.com/k1LoW/curlreq"
	"gopkg.in/yaml.v2"
)

type runbook struct {
	Desc    string          `yaml:"desc"`
	Runners yaml.MapSlice   `yaml:"runners,omitempty"`
	Steps   []yaml.MapSlice `yaml:"steps"`
}

func NewRunbook(desc string) *runbook {
	r := &runbook{Desc: desc}
	return r
}

func (rb *runbook) AppendStep(in ...string) error {
	if len(in) == 0 {
		return errors.New("no argument")
	}
	switch {
	case strings.HasPrefix(in[0], "curl"):
		return rb.curlToStep(in...)
	default:
		return rb.cmdToStep(in...)
	}
}

func (rb *runbook) curlToStep(in ...string) error {
	req, err := curlreq.NewRequest(in...)
	if err != nil {
		return err
	}

	splitted := strings.Split(req.URL.String(), req.URL.Host)
	dsn := fmt.Sprintf("%s%s", splitted[0], req.URL.Host)
	key := rb.setRunner(dsn)
	step, err := CreateHTTPStepMapSlice(key, req)
	if err != nil {
		return err
	}
	rb.Steps = append(rb.Steps, step)
	return nil
}

func (rb *runbook) setRunner(dsn string) string {
	const (
		httpRunnerKeyPrefix = "req"
		grpcRunnerKeyPrefix = "greq"
		dbRunnerKeyPrefix   = "db"
	)
	var hc, gc, dc int
	for _, r := range rb.Runners {
		v := r.Value.(string)
		if v == dsn {
			return r.Key.(string)
		}
		switch {
		case strings.HasPrefix(v, "http"):
			hc += 1
		case strings.HasPrefix(v, "grpc"):
			gc += 1
		default:
			dc += 1
		}
	}

	var key string
	switch {
	case strings.HasPrefix(dsn, "http"):
		if hc > 0 {
			key = fmt.Sprintf("%s%d", httpRunnerKeyPrefix, hc+1)
		} else {
			key = httpRunnerKeyPrefix
		}
	case strings.HasPrefix(dsn, "grpc"):
		if gc > 0 {
			key = fmt.Sprintf("%s%d", grpcRunnerKeyPrefix, gc+1)
		} else {
			key = grpcRunnerKeyPrefix
		}
	default:
		if dc > 0 {
			key = fmt.Sprintf("%s%d", dbRunnerKeyPrefix, dc+1)
		} else {
			key = dbRunnerKeyPrefix
		}
	}
	rb.Runners = append(rb.Runners, yaml.MapItem{Key: key, Value: dsn})
	return key
}

func (rb *runbook) cmdToStep(in ...string) error {
	step := yaml.MapSlice{
		{Key: execRunnerKey, Value: yaml.MapSlice{
			{Key: "command", Value: joinCommands(in...)},
		}},
	}
	rb.Steps = append(rb.Steps, step)
	return nil
}

func joinCommands(in ...string) string {
	var cmd []string
	for _, i := range in {
		i = strings.TrimSuffix(i, "\n")
		if strings.Contains(i, " ") {
			cmd = append(cmd, fmt.Sprintf("%#v", i))
		} else {
			cmd = append(cmd, i)
		}
	}
	return strings.Join(cmd, " ") + "\n"
}
