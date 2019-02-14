package readline

import (
	"bufio"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

type Terminal struct {
	cfg       *Config
	outchan   chan rune
	closed    int32
	stopChan  chan struct{}
	kickChan  chan struct{}
	wg        sync.WaitGroup
	isReading int32
	sleeping  int32
}

func NewTerminal(cfg *Config) (*Terminal, error) {
	if err := cfg.Init(); err != nil {
		return nil, err
	}
	t := &Terminal{
		cfg:      cfg,
		kickChan: make(chan struct{}, 1),
		outchan:  make(chan rune),
		stopChan: make(chan struct{}, 1),
	}

	go t.ioloop()
	return t, nil
}

// SleepToResume will sleep myself, and return only if I'm resumed.
func (t *Terminal) SleepToResume() {
	if !atomic.CompareAndSwapInt32(&t.sleeping, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&t.sleeping, 0)

	t.ExitRawMode()
	ch := WaitForResume()
	SuspendMe()
	<-ch
	t.EnterRawMode()
}

func (t *Terminal) EnterRawMode() (err error) {
	return t.cfg.FuncMakeRaw()
}

func (t *Terminal) ExitRawMode() (err error) {
	return t.cfg.FuncExitRaw()
}

func (t *Terminal) Write(b []byte) (int, error) {
	return t.cfg.Stdout.Write(b)
}

func (t *Terminal) Print(s string) {
	fmt.Fprintf(t.cfg.Stdout, "%s", s)
}

func (t *Terminal) PrintRune(r rune) {
	fmt.Fprintf(t.cfg.Stdout, "%c", r)
}

func (t *Terminal) Readline() *Operation {
	return NewOperation(t, t.cfg)
}

// return rune(0) if meet EOF
func (t *Terminal) ReadRune() rune {
	ch, ok := <-t.outchan
	if !ok {
		return rune(0)
	}
	return ch
}

func (t *Terminal) IsReading() bool {
	return atomic.LoadInt32(&t.isReading) == 1
}

func (t *Terminal) KickRead() {
	select {
	case t.kickChan <- struct{}{}:
	default:
	}
}

func (t *Terminal) ioloop() {
	t.wg.Add(1)
	defer t.wg.Done()
	var (
		isEscape       bool
		isEscapeEx     bool
		expectNextChar bool
	)

	buf := bufio.NewReader(t.cfg.Stdin)
	for {
		if !expectNextChar {
			atomic.StoreInt32(&t.isReading, 0)
			select {
			case <-t.kickChan:
				atomic.StoreInt32(&t.isReading, 1)
			case <-t.stopChan:
				return
			}
		}
		expectNextChar = false
		r, _, err := buf.ReadRune()
		if err != nil {
			if strings.Contains(err.Error(), "interrupted system call") {
				expectNextChar = true
				continue
			}
			break
		}

		if isEscape {
			isEscape = false
			if r == CharEscapeEx {
				expectNextChar = true
				isEscapeEx = true
				continue
			}
			r = escapeKey(r, buf)
		} else if isEscapeEx {
			isEscapeEx = false
			r = escapeExKey(r, buf)
		}

		expectNextChar = true
		switch r {
		case CharEsc:
			if t.cfg.VimMode {
				t.outchan <- r
				break
			}
			isEscape = true
		case CharInterrupt, CharEnter, CharCtrlJ, CharDelete:
			expectNextChar = false
			fallthrough
		default:
			t.outchan <- r
		}
	}
	close(t.outchan)
}

func (t *Terminal) Bell() {
	fmt.Fprintf(t, "%c", CharBell)
}

func (t *Terminal) Close() error {
	if atomic.SwapInt32(&t.closed, 1) != 0 {
		return nil
	}
	t.stopChan <- struct{}{}
	t.wg.Wait()
	return t.ExitRawMode()
}

func (t *Terminal) SetConfig(c *Config) error {
	if err := c.Init(); err != nil {
		return err
	}
	t.cfg = c
	return nil
}
