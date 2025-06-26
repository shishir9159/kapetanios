package main

import (
	"bytes"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os/exec"
	"strconv"
)

type JournalOptions struct {
	Follow     bool
	NumEntries int
	Since      string
	UnitName   string
	FullOutput bool
}

func getJournalLogs(opts JournalOptions) (string, error) {
	var args []string

	args = append(args, "-u", opts.UnitName)

	if opts.Follow {
		args = append(args, "-f")
	}

	if opts.NumEntries > 0 {
		args = append(args, "-n", strconv.Itoa(opts.NumEntries))
	}

	if opts.Since != "" {
		args = append(args, "--since", opts.Since)
	}

	if opts.FullOutput {
		args = append(args, "-x")
	}

	cmd := exec.Command("journalctl", args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get logs for service %s: %w", opts.UnitName, err)
	}

	return stdoutBuf.String(), nil
}

//    - `-f` for following logs
//    - `-n` for limiting number of entries
//    - `--since` for time-based filtering
//    - for full output with explanatory messages `-x`

func main() {

	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "message",
			LevelKey:     "level",
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			TimeKey:      "time",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	logger := zap.Must(cfg.Build())
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			logger.Info("error syncing logger before application terminates",
				zap.Error(er))
		}
	}(logger)

	journalConfig := JournalOptions{
		Follow:     false,
		NumEntries: 100,
		Since:      "1 hour ago",
		UnitName:   "kubelet",
		FullOutput: true,
	}

	logs, err := getJournalLogs(journalConfig)
	if err != nil {
		logger.Error("error getting journal logs",
			zap.Error(err))
	}

	logger.Info("scrapped logs",
		zap.String("logs", logs))
}
