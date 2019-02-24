package events_stream

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/bettercap/bettercap/session"

	"github.com/evilsocket/islazy/str"
)

var (
	ErrEmptyExpression = errors.New("expression can not be empty")
)

type IgnoreFilter string

func (f IgnoreFilter) Matches(s string) bool {
	return string(f) == s || strings.HasPrefix(s, string(f))
}

type IgnoreList struct {
	sync.RWMutex
	filters []IgnoreFilter
}

func NewIgnoreList() *IgnoreList {
	return &IgnoreList{
		filters: make([]IgnoreFilter, 0),
	}
}

func (l *IgnoreList) checkExpression(expr string) (string, error) {
	expr = str.Trim(expr)
	if expr == "" {
		return "", ErrEmptyExpression
	}

	return expr, nil
}

func (l *IgnoreList) Add(expr string) (err error) {
	if expr, err = l.checkExpression(expr); err != nil {
		return err
	}

	l.Lock()
	defer l.Unlock()

	// first check for duplicates
	for _, filter := range l.filters {
		if filter.Matches(expr) {
			return fmt.Errorf("filter '%s' already matches the expression '%s'", filter, expr)
		}
	}

	// all good
	l.filters = append(l.filters, IgnoreFilter(expr))

	return nil
}

func (l *IgnoreList) Remove(expr string) (err error) {
	if expr, err = l.checkExpression(expr); err != nil {
		return err
	}

	l.Lock()
	defer l.Unlock()

	// build a new list with everything that does not match
	toRemove := IgnoreFilter(expr)
	newList := make([]IgnoreFilter, 0)
	for _, filter := range l.filters {
		if !toRemove.Matches(string(filter)) {
			newList = append(newList, filter)
		}
	}

	if len(newList) == len(l.filters) {
		return fmt.Errorf("expression '%s' did not match any filter", expr)
	}

	// swap
	l.filters = newList

	return nil
}

func (l *IgnoreList) Ignored(e session.Event) bool {
	l.RLock()
	defer l.RUnlock()

	for _, filter := range l.filters {
		if filter.Matches(e.Tag) {
			return true
		}
	}

	return false
}

func (l *IgnoreList) Empty() bool {
	l.RLock()
	defer l.RUnlock()
	return len(l.filters) == 0
}

func (l *IgnoreList) Filters() []IgnoreFilter {
	return l.filters
}
