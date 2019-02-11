package utils

import (
	"fmt"

	"regexp"
	"strings"

	"github.com/bettercap/bettercap/session"

	"github.com/evilsocket/islazy/tui"
)

type ViewSelector struct {
	owner *session.SessionModule

	Filter     string
	filterName string
	filterPrev string
	Expression *regexp.Regexp

	SortField  string
	Sort       string
	SortSymbol string
	sortFields map[string]bool
	sortName   string
	sortParser string
	sortParse  *regexp.Regexp

	Limit     int
	limitName string
}

func ViewSelectorFor(m *session.SessionModule, prefix string, sortFields []string, defExpression string) *ViewSelector {
	parser := "(" + strings.Join(sortFields, "|") + ") (desc|asc)"
	s := &ViewSelector{
		owner:      m,
		filterName: prefix + ".filter",
		sortName:   prefix + ".sort",
		sortParser: parser,
		sortParse:  regexp.MustCompile(parser),
		limitName:  prefix + ".limit",
	}

	m.AddParam(session.NewStringParameter(s.filterName, "", "", "Defines a regular expression filter for "+prefix))
	m.AddParam(session.NewStringParameter(
		s.sortName,
		defExpression,
		s.sortParser,
		"Defines sorting field ("+strings.Join(sortFields, ", ")+") and direction (asc or desc) for "+prefix))

	m.AddParam(session.NewIntParameter(s.limitName, "0", "Defines limit for "+prefix))

	s.parseSorting()

	return s
}

func (s *ViewSelector) parseFilter() (err error) {
	if err, s.Filter = s.owner.StringParam(s.filterName); err != nil {
		return
	}

	if s.Filter != "" {
		if s.Filter != s.filterPrev {
			if s.Expression, err = regexp.Compile(s.Filter); err != nil {
				return
			}
		}
	} else {
		s.Expression = nil
	}
	s.filterPrev = s.Filter
	return
}

func (s *ViewSelector) parseSorting() (err error) {
	expr := ""
	if err, expr = s.owner.StringParam(s.sortName); err != nil {
		return
	}

	tokens := s.sortParse.FindAllStringSubmatch(expr, -1)
	if tokens == nil {
		return fmt.Errorf("expression '%s' doesn't parse", expr)
	}

	s.SortField = tokens[0][1]
	s.Sort = tokens[0][2]
	s.SortSymbol = tui.Blue("▾")
	if s.Sort == "asc" {
		s.SortSymbol = tui.Blue("▴")
	}

	return
}

func (s *ViewSelector) Update() (err error) {
	if err = s.parseFilter(); err != nil {
		return
	} else if err = s.parseSorting(); err != nil {
		return
	} else if err, s.Limit = s.owner.IntParam(s.limitName); err != nil {
		return
	}
	return
}
