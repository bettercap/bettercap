package tld

import "testing"

func run(input, sub, dom, tld string, t *testing.T) {

	u, err := Parse(input)

	if err != nil {
		t.Errorf("errored '%s'", err)
	} else if u.TLD != tld {
		t.Errorf("should have TLD '%s', got '%s'", tld, u.TLD)
	} else if u.Domain != dom {
		t.Errorf("should have Domain '%s', got '%s'", dom, u.Domain)
	} else if u.Subdomain != sub {
		t.Errorf("should have Subdomain '%s', got '%s'", sub, u.Subdomain)
	}
}

func Test0(t *testing.T) {
	run("http://foo.com", "", "foo", "com", t)
}

func Test1(t *testing.T) {
	run("http://zip.zop.foo.com", "zip.zop", "foo", "com", t)
}

func Test2(t *testing.T) {
	run("http://au.com.au", "", "au", "com.au", t)
}

func Test3(t *testing.T) {
	run("http://im.from.england.co.uk:1900", "im.from", "england", "co.uk", t)
}
