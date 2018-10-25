package session

import (
	"fmt"
	"testing"
)

func TestParseCommands(t *testing.T) {
	//commands := ParseCommands("wifi.recon on; asdf; \"asdf;\" asdf")
	t.Run("handles a semicolon as a delimiter", func(t *testing.T) {
		first := "wifi.recon on"
		second := "wifi.ap"
		cmd := fmt.Sprintf("%s; %s", first, second)
		commands := ParseCommands(cmd)
		if l := len(commands); l != 2 {
			t.Fatalf("Expected 2 commands, got %d", l)
		}
		if got := commands[0]; got != first {
			t.Fatalf("expected %s got %s", first, got)
		}
		if got := commands[1]; got != second {
			t.Fatalf("expected %s got %s", second, got)
		}
	})
	t.Run("handles semicolon inside quotes", func(t *testing.T) {
		cmd := "set ticker.commands \"clear; net.show\""
		commands := ParseCommands(cmd)
		if l := len(commands); l != 1 {
			t.Fatalf("expected 1 command, got %d", l)
		}
		// Expect double-quotes stripped
		expected := "set ticker.commands clear; net.show"
		if got := commands[0]; got != expected {
			fmt.Println(got)
			t.Fatalf("expected %s got %s", cmd, got)
		}
	})
	t.Run("handles semicolon inside single quotes", func(t *testing.T) {
		cmd := "set ticker.commands 'clear; net.show'"
		commands := ParseCommands(cmd)
		if l := len(commands); l != 1 {
			t.Fatalf("expected 1 command, got %d", l)
		}
		// Expect double-quotes stripped
		expected := "set ticker.commands clear; net.show"
		if got := commands[0]; got != expected {
			fmt.Println(got)
			t.Fatalf("expected %s got %s", cmd, got)
		}
	})
	t.Run("handles semicolon inside single quotes inside quote", func(t *testing.T) {
		cmd := "set ticker.commands \"'clear; net.show'\""
		commands := ParseCommands(cmd)
		if l := len(commands); l != 1 {
			t.Fatalf("expected 1 command, got %d", l)
		}
		// Expect double-quotes stripped
		expected := "set ticker.commands 'clear; net.show'"
		if got := commands[0]; got != expected {
			fmt.Println(got)
			t.Fatalf("expected %s got %s", cmd, got)
		}
	})
	t.Run("handles semicolon inside quotes inside single quote", func(t *testing.T) {
		cmd := "set ticker.commands '\"clear; net.show\"'"
		commands := ParseCommands(cmd)
		if l := len(commands); l != 1 {
			t.Fatalf("expected 1 command, got %d", l)
		}
		// Expect double-quotes stripped
		expected := "set ticker.commands \"clear; net.show\""
		if got := commands[0]; got != expected {
			fmt.Println(got)
			t.Fatalf("expected %s got %s", cmd, got)
		}
	})
	t.Run("handle mismatching quote", func(t *testing.T) {
		cmd := "set ticker.commands \"clear; echo it's working ?\""
		commands := ParseCommands(cmd)
		if l := len(commands); l != 1 {
			t.Fatalf("expected 1 command, got %d", l)
		}
		// Expect double-quotes stripped
		expected := "set ticker.commands clear; echo it's working ?"
		if got := commands[0]; got != expected {
			fmt.Println(got)
			t.Fatalf("expected %s got %s", cmd, got)
		}
	})
}
