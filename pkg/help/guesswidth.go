// +build appengine !linux,!freebsd,!darwin,!dragonfly,!netbsd,!openbsd

package help

import "io"

func guessWidth(w io.Writer) int {
	return 120
}
