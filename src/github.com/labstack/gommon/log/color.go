//go:build !appengine
// +build !appengine

package log

import (
	"io"
)

func output() io.Writer {
	return colorable.NewColorableStdout()
}
