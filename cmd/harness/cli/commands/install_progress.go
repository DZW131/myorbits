package commands

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type installProgressMode string

const (
	installProgressAuto  installProgressMode = "auto"
	installProgressPlain installProgressMode = "plain"
	installProgressQuiet installProgressMode = "quiet"
)

type installProgressEmitter struct {
	writer  io.Writer
	enabled bool
}

func newInstallProgressEmitter(writer io.Writer, rawMode string) (installProgressEmitter, error) {
	mode, err := parseInstallProgressMode(rawMode)
	if err != nil {
		return installProgressEmitter{}, err
	}

	switch mode {
	case installProgressPlain:
		return installProgressEmitter{writer: writer, enabled: true}, nil
	case installProgressQuiet:
		return installProgressEmitter{writer: writer, enabled: false}, nil
	case installProgressAuto:
		return installProgressEmitter{writer: writer, enabled: writerIsTerminal(writer)}, nil
	default:
		return installProgressEmitter{}, fmt.Errorf("unsupported install progress mode %q", rawMode)
	}
}

func parseInstallProgressMode(rawMode string) (installProgressMode, error) {
	trimmed := strings.TrimSpace(rawMode)
	if trimmed == "" {
		return installProgressAuto, nil
	}

	switch installProgressMode(trimmed) {
	case installProgressAuto, installProgressPlain, installProgressQuiet:
		return installProgressMode(trimmed), nil
	default:
		return "", fmt.Errorf("invalid --progress value %q (expected auto, plain, or quiet)", rawMode)
	}
}

func writerIsTerminal(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	if !ok {
		return false
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}

func (emitter installProgressEmitter) Stage(stage string) error {
	if !emitter.enabled {
		return nil
	}
	if _, err := fmt.Fprintf(emitter.writer, "progress: %s\n", strings.TrimSpace(stage)); err != nil {
		return fmt.Errorf("write install progress: %w", err)
	}
	return nil
}
