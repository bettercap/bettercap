//go:generate sh generate.sh

//Package tld has the same API as net/url except
//tld.URL contains extra fields: Subdomain, Domain, TLD and Port.
package tld

import (
	"errors"
	"net/url"
)

//URL embeds net/url and adds extra fields ontop
type URL struct {
	Subdomain, Domain, TLD, Port string
	*url.URL
}

//Parse mirrors net/url.Parse except instead it returns
//a tld.URL, which contains extra fields.
func Parse(s string) (*URL, error) {

	url, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	if url.Host == "" {
		return &URL{URL: url}, nil
	}

	dom, port := domainPort(url.Host)
	//index of tld
	tld := 0
	i := 0
	l := len(dom) - 1

	//binary search the TLD list
	lo := 0
	hi := count - 1
	for lo != hi && lo+1 != hi {

		mid := (hi + lo) / 2
		guess := list[mid]

		//for binary search debugging...
		// log.Printf("[%d - %d - %d] %s == %s (%s)", lo, mid, hi, string(dom[l-i]), string(guess[i]), guess)

		if i < len(guess) && i <= l && guess[i] == dom[l-i] {
			//store partial match
			if i > tld && guess[i] == '.' {
				tld = i
			}
			//advance!
			i++
			//checked all is in guess
			if len(guess) == i && dom[l-i] == '.' {
				tld = i
				break
			}
		} else if i >= len(guess) || (i <= l && guess[i] < dom[l-i]) {
			lo = mid
			i = 0
		} else {
			hi = mid
			i = 0
		}
	}

	if tld == 0 {
		return nil, errors.New("tld not found")
	}

	//extract the tld
	t := dom[l-tld+1:]
	//we can calculate the root domain
	dom = dom[:l-tld]
	//and subdomain
	sub := ""
	for i := len(dom) - 1; i >= 0; i-- {
		if dom[i] == '.' {
			sub = dom[:i]
			dom = dom[i+1:]
			break
		}
	}

	return &URL{
		Subdomain: sub,
		Domain:    dom,
		TLD:       t,
		Port:      port,
		URL:       url,
	}, nil
}

func domainPort(host string) (string, string) {
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			return host[:i], host[i+1:]
		} else if host[i] < '0' || host[i] > '9' {
			return host, ""
		}
	}
	//will only land here if the string is all digits,
	//net/url should prevent that from happening
	return host, ""
}
