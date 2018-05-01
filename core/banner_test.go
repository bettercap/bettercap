package core

import (
	"regexp"
	"testing"
)

func TestBannerName(t *testing.T) {
	if Name != "bettercap" {
		t.Fatalf("expected '%s', got '%s'", "bettercap", Name)
	}
}
func TestBannerWebsite(t *testing.T) {
	if Website != "https://bettercap.org/" {
		t.Fatalf("expected '%s', got '%s'", "https://bettercap.org/", Website)
	}
}

func TestBannerVersion(t *testing.T) {
	match, err := regexp.MatchString(`\d+.\d+`, Version)
	if err != nil {
		t.Fatalf("unable to perform regex on Version constant")
	}
	if !match {
		t.Fatalf("expected Version constant in format '%s', got '%s'", "X.X", Version)
	}
}

func TestBannerAuthor(t *testing.T) {
	if Author != "Simone 'evilsocket' Margaritelli" {
		t.Fatalf("expected '%s', got '%s'", "Simone 'evilsocket' Margaritelli", Author)
	}
}
