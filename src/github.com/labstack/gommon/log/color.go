//go:build !appengine
// +build !appengine

package log

import (
	"io"

	"github.com/davy66666/poker-go/src/github.com/mattn/go-colorable"
)

func output() io.Writer {
	return colorable.NewColorableStdout()
}
