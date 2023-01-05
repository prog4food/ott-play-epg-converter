package helpers

import (
	"io"
	"os"
	"runtime"

	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func getStdOut() io.Writer {
	// Фикс цветной консоли для Windows
	if runtime.GOOS == "windows" {
		return colorable.NewColorableStdout()
	}
	return os.Stdout
}

func getStdErr() io.Writer {
	// Фикс цветной консоли для Windows
	if runtime.GOOS == "windows" {
		return colorable.NewColorableStderr()
	}
	return os.Stderr
}

func InitLogger(msg1, msg2 string, stdErr bool) {
	stdOutFunc := getStdOut
	if stdErr {
		stdOutFunc = getStdErr
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        stdOutFunc(),
		TimeFormat: "2006-01-02T15:04:05",
	})
	log.Info().Msg(msg1 + msg2)
	log.Info().Msg("  git@prog4food (c) 2o22")
}
