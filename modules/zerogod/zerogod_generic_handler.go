package zerogod

import (
	"fmt"

	"github.com/evilsocket/islazy/tui"
)

func Dump(by []byte) string {
	s := ""
	n := len(by)
	rowcount := 0
	width := 16

	stop := (n / width) * width
	k := 0
	for i := 0; i <= stop; i += width {
		k++
		if i+width < n {
			rowcount = width
		} else {
			rowcount = min(k*width, n) % width
		}

		s += fmt.Sprintf("%02d ", i)
		for j := 0; j < rowcount; j++ {
			s += fmt.Sprintf("%02x  ", by[i+j])
		}
		for j := rowcount; j < width; j++ {
			s += "    "
		}
		s += fmt.Sprintf("  '%s'\n", viewString(by[i:(i+rowcount)]))
	}

	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func viewString(b []byte) string {
	r := []rune(string(b))
	for i := range r {
		if r[i] < 32 || r[i] > 126 {
			r[i] = '.'
		}
	}
	return string(r)
}

func handleGenericTCP(ctx *HandlerContext) {
	defer ctx.client.Close()

	ctx.mod.Info("accepted generic tcp connection for service %s (port %d): %v", tui.Green(ctx.service), ctx.srvPort, ctx.client.RemoteAddr())

	buf := make([]byte, 1024)
	for {
		if read, err := ctx.client.Read(buf); err != nil {
			ctx.mod.Error("error while reading from %v: %v", ctx.client.RemoteAddr(), err)
			break
		} else if read == 0 {
			ctx.mod.Error("error while reading from %v: no data", ctx.client.RemoteAddr())
			break
		} else {
			ctx.mod.Info("read %d bytes from %v:\n%s\n", read, ctx.client.RemoteAddr(), Dump(buf[0:read]))
		}
	}
}
