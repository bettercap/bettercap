package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/evilsocket/islazy/str"
)

var (
	ErrEmptyExpression = errors.New("expression can not be empty")
)

type filter string

func (f filter) Matches(s string) bool {
	return string(f) == s || strings.HasPrefix(s, string(f))
}

type EventsIgnoreList struct {
	sync.RWMutex
	filters []filter
}

func NewEventsIgnoreList() *EventsIgnoreList {
	return &EventsIgnoreList{
		filters: make([]filter, 0),
	}
}

func (l *EventsIgnoreList) MarshalJSON() ([]byte, error) {
	l.RLock()
	defer l.RUnlock()
	return json.Marshal(l.filters)
}

func (l *EventsIgnoreList) checkExpression(expr string) (string, error) {
	expr = str.Trim(expr)
	if expr == "" {
		return "", ErrEmptyExpression
	}

	return expr, nil
}

func (l *EventsIgnoreList) Add(expr string) (err error) {
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
	l.filters = append(l.filters, filter(expr))

	return nil
}

func (l *EventsIgnoreList) Remove(expr string) (err error) {
	if expr, err = l.checkExpression(expr); err != nil {
		return err
	}

	l.Lock()
	defer l.Unlock()

	// build a new list with everything that does not match
	toRemove := filter(expr)
	newList := make([]filter, 0)
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

func (l *EventsIgnoreList) Clear() {
	l.Lock()
	defer l.Unlock()
	l.filters = make([]filter, 0)
}

func (l *EventsIgnoreList) Ignored(e Event) bool {
	l.RLock()
	defer l.RUnlock()

	for _, filter := range l.filters {
		if filter.Matches(e.Tag) {
			return true
		}
	}

	return false
}

func (l *EventsIgnoreList) Empty() bool {
	l.RLock()
	defer l.RUnlock()
	return len(l.filters) == 0
}

func (l *EventsIgnoreList) Filters() []filter {
	return l.filters
}
