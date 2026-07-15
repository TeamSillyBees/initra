package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const doctorCommandTimeout = 15 * time.Second

type doctorOptions struct {
	json bool
}

type doctorStatus string

const (
	doctorStatusOK       doctorStatus = "OK"
	doctorStatusBroken   doctorStatus = "BROKEN"
	doctorStatusMissing  doctorStatus = "MISSING"
	doctorStatusOptional doctorStatus = "OPTIONAL"
)

type doctorCheck struct {
	Name     string       `json:"name"`
	Status   doctorStatus `json:"status"`
	Required bool         `json:"required"`
	Detail   string       `json:"detail"`
}

type doctorReport struct {
	SchemaVersion int           `json:"schema_version"`
	OK            bool          `json:"ok"`
	Checks        []doctorCheck `json:"checks"`
}

func newDoctorCommand(stdout io.Writer) *cobra.Command {
	opts := doctorOptions{}
	cmd := &cobra.Command{
		Use:           "doctor",
		Short:         "检查本地开发环境",
		Long:          "检查当前开发环境中 Go、Atlas、Ent、golangci-lint 和标准项目配置文件是否可用。",
		Example:       "  initra doctor\n  initra doctor --json",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          requireNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctorChecks(cmd.OutOrStdout(), opts)
		},
	}
	configureCommand(cmd, stdout)
	cmd.Flags().BoolVar(&opts.json, "json", false, "输出稳定的 JSON 检查报告")
	return cmd
}

func runDoctorChecks(stdout io.Writer, opts doctorOptions) error {
	if stdout == nil {
		stdout = io.Discard
	}
	report := collectDoctorReport()
	if opts.json {
		encoder := json.NewEncoder(stdout)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(report); err != nil {
			return fmt.Errorf("输出 doctor JSON 失败: %w", err)
		}
	} else {
		for _, check := range report.Checks {
			if _, err := fmt.Fprintf(stdout, "%s: %s %s\n", check.Name, check.Status, check.Detail); err != nil {
				return fmt.Errorf("输出 doctor 结果失败: %w", err)
			}
		}
	}
	if report.OK {
		return nil
	}
	failed := 0
	for _, check := range report.Checks {
		if check.Required && check.Status != doctorStatusOK {
			failed++
		}
	}
	return fmt.Errorf("doctor 检查失败：%d 个必需项不可用", failed)
}

func collectDoctorReport() doctorReport {
	checks := []doctorCheck{
		checkDoctorTool("Go", true, "go", "version"),
		checkDoctorTool("Atlas", true, "atlas", "version"),
		checkDoctorTool("Ent", true, "go", "tool", "ent", "--help"),
		checkDoctorTool("golangci-lint", true, "golangci-lint", "version"),
		checkDoctorFile("config.yaml", false, filepath.Join("configs", "config.yaml")),
		checkDoctorFile("config.dev.yaml", false, filepath.Join("configs", "config.dev.yaml")),
		checkDoctorFile("Atlas config", true, filepath.Join("db", "atlas.hcl")),
	}
	report := doctorReport{SchemaVersion: 1, OK: true, Checks: checks}
	for _, check := range checks {
		if check.Required && check.Status != doctorStatusOK {
			report.OK = false
			break
		}
	}
	return report
}

func checkDoctorTool(name string, required bool, command string, args ...string) doctorCheck {
	return checkDoctorToolWithTimeout(name, required, doctorCommandTimeout, command, args...)
}

func checkDoctorToolWithTimeout(name string, required bool, timeout time.Duration, command string, args ...string) doctorCheck {
	path, err := exec.LookPath(command)
	if err != nil {
		return doctorCheck{Name: name, Status: doctorStatusMissing, Required: required, Detail: command}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	output, err := exec.CommandContext(ctx, command, args...).CombinedOutput()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return doctorCheck{
				Name:     name,
				Status:   doctorStatusBroken,
				Required: required,
				Detail:   fmt.Sprintf("timeout after %s", timeout),
			}
		}
		detail := firstLine(string(output))
		if detail == "" {
			detail = err.Error()
		}
		return doctorCheck{Name: name, Status: doctorStatusBroken, Required: required, Detail: detail}
	}
	detail := firstLine(string(output))
	if detail == "" {
		detail = path
	}
	return doctorCheck{Name: name, Status: doctorStatusOK, Required: required, Detail: detail}
}

func checkDoctorFile(name string, required bool, path string) doctorCheck {
	info, err := os.Stat(path)
	switch {
	case err == nil && !info.IsDir():
		return doctorCheck{Name: name, Status: doctorStatusOK, Required: required, Detail: path}
	case err == nil:
		return doctorCheck{Name: name, Status: doctorStatusBroken, Required: required, Detail: path + " is a directory"}
	case os.IsNotExist(err) && !required:
		return doctorCheck{Name: name, Status: doctorStatusOptional, Required: false, Detail: path + " missing"}
	case os.IsNotExist(err):
		return doctorCheck{Name: name, Status: doctorStatusMissing, Required: true, Detail: path}
	default:
		return doctorCheck{Name: name, Status: doctorStatusBroken, Required: required, Detail: err.Error()}
	}
}

func firstLine(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if index := strings.IndexAny(value, "\r\n"); index >= 0 {
		return value[:index]
	}
	return value
}
