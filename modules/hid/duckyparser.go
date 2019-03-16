package hid

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/evilsocket/islazy/fs"
)

type DuckyParser struct {
	mod *HIDRecon
}

func (p DuckyParser) parseLiteral(what string, kmap KeyMap) (*Command, error) {
	// get reference command from the layout
	ref, found := kmap[what]
	if found == false {
		return nil, fmt.Errorf("can't find '%s' in current keymap", what)
	}
	return &Command{
		HID:  ref.HID,
		Mode: ref.Mode,
	}, nil
}

func (p DuckyParser) parseModifier(line string, kmap KeyMap, modMask byte) (*Command, error) {
	// get optional key after the modifier
	ch := ""
	if idx := strings.IndexRune(line, ' '); idx != -1 {
		ch = line[idx+1:]
	}
	cmd, err := p.parseLiteral(ch, kmap)
	if err != nil {
		return nil, err
	}
	// apply modifier mask
	cmd.Mode |= modMask
	return cmd, nil
}

func (p DuckyParser) parseNumber(from string) (int, error) {
	idx := strings.IndexRune(from, ' ')
	if idx == -1 {
		return 0, fmt.Errorf("can't parse number from '%s'", from)
	}

	num, err := strconv.Atoi(from[idx+1:])
	if err != nil {
		return 0, fmt.Errorf("can't parse number from '%s': %v", from, err)
	}

	return num, nil
}

func (p DuckyParser) parseString(from string) (string, error) {
	idx := strings.IndexRune(from, ' ')
	if idx == -1 {
		return "", fmt.Errorf("can't parse string from '%s'", from)
	}
	return from[idx+1:], nil
}

func (p DuckyParser) lineIs(line string, tokens ...string) bool {
	for _, tok := range tokens {
		if strings.HasPrefix(line, tok) {
			return true
		}
	}
	return false
}

func (p DuckyParser) Parse(kmap KeyMap, path string) (cmds []*Command, err error) {
	lines := []string{}
	source := []string{}
	reader := (chan string)(nil)

	if reader, err = fs.LineReader(path); err != nil {
		return
	} else {
		for line := range reader {
			lines = append(lines, line)
		}
	}

	// preprocessing
	for lineno, line := range lines {
		if p.lineIs(line, "REPEAT") {
			if lineno == 0 {
				err = fmt.Errorf("error on line %d: REPEAT instruction at the beginning of the script", lineno+1)
				return
			}
			times := 1
			times, err = p.parseNumber(line)
			if err != nil {
				return
			}

			for i := 0; i < times; i++ {
				source = append(source, lines[lineno-1])
			}
		} else {
			source = append(source, line)
		}
	}

	cmds = make([]*Command, 0)
	for _, line := range source {
		cmd := &Command{}
		if p.lineIs(line, "CTRL", "CONTROL") {
			if cmd, err = p.parseModifier(line, kmap, 1); err != nil {
				return
			}
		} else if p.lineIs(line, "SHIFT") {
			if cmd, err = p.parseModifier(line, kmap, 2); err != nil {
				return
			}
		} else if p.lineIs(line, "ALT") {
			if cmd, err = p.parseModifier(line, kmap, 4); err != nil {
				return
			}
		} else if p.lineIs(line, "GUI", "WINDOWS", "COMMAND") {
			if cmd, err = p.parseModifier(line, kmap, 8); err != nil {
				return
			}
		} else if p.lineIs(line, "CTRL-ALT", "CONTROL-ALT") {
			if cmd, err = p.parseModifier(line, kmap, 4|1); err != nil {
				return
			}
		} else if p.lineIs(line, "CTRL-SHIFT", "CONTROL-SHIFT") {
			if cmd, err = p.parseModifier(line, kmap, 1|2); err != nil {
				return
			}
		} else if p.lineIs(line, "ESC", "ESCAPE", "APP") {
			if cmd, err = p.parseLiteral("ESCAPE", kmap); err != nil {
				return
			}
		} else if p.lineIs(line, "ENTER") {
			if cmd, err = p.parseLiteral("ENTER", kmap); err != nil {
				return
			}
		} else if p.lineIs(line, "UP", "UPARROW") {
			if cmd, err = p.parseLiteral("UP", kmap); err != nil {
				return
			}
		} else if p.lineIs(line, "DOWN", "DOWNARROW") {
			if cmd, err = p.parseLiteral("DOWN", kmap); err != nil {
				return
			}
		} else if p.lineIs(line, "LEFT", "LEFTARROW") {
			if cmd, err = p.parseLiteral("LEFT", kmap); err != nil {
				return
			}
		} else if p.lineIs(line, "RIGHT", "RIGHTARROW") {
			if cmd, err = p.parseLiteral("RIGHT", kmap); err != nil {
				return
			}
		} else if p.lineIs(line, "DELAY", "SLEEP") {
			secs := 0
			if secs, err = p.parseNumber(line); err != nil {
				return
			}
			cmd = &Command{Sleep: secs}
		} else if p.lineIs(line, "STRING", "STR") {
			str := ""
			if str, err = p.parseString(line); err != nil {
				return
			}

			for _, c := range str {
				if cmd, err = p.parseLiteral(string(c), kmap); err != nil {
					return
				}
				cmds = append(cmds, cmd)
			}

			continue
		} else if cmd, err = p.parseLiteral(line, kmap); err != nil {
			err = fmt.Errorf("error parsing '%s': %s", line, err)
			return
		}

		cmds = append(cmds, cmd)
	}

	return
}
