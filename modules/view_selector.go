package modules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bettercap/bettercap/session"
)

type ViewSelector struct {
	owner *session.SessionModule

	Filter     string
	filterName string
	filterPrev string
	Expression *regexp.Regexp

	SortBy     string
	sortBys    map[string]bool
	sortByName string

	Sort     string
	sortName string

	Limit     int
	limitName string
}

func ViewSelectorFor(m *session.SessionModule, prefix string, sortBys []string, defSortBy string) *ViewSelector {
	s := &ViewSelector{
		owner:      m,
		filterName: prefix + ".filter",
		filterPrev: "",
		sortByName: prefix + ".sort_by",
		sortBys:    make(map[string]bool),
		sortName:   prefix + ".sort",
		limitName:  prefix + ".limit",

		SortBy:     defSortBy,
		Sort:       "asc",
		Limit:      0,
		Filter:     "",
		Expression: nil,
	}

	for _, sb := range sortBys {
		s.sortBys[sb] = true
	}

	m.AddParam(session.NewStringParameter(s.filterName, "", "", "Defines a regular expression filter for "+prefix))
	m.AddParam(session.NewStringParameter(s.sortByName, defSortBy, "", "Defines sorting field for "+prefix+", available: "+strings.Join(sortBys, ", ")))
	m.AddParam(session.NewStringParameter(s.sortName, "asc", "", "Defines sorting direction for "+prefix))
	m.AddParam(session.NewIntParameter(s.limitName, "0", "Defines limit for "+prefix))
	return s
}

func (s *ViewSelector) Update() (err error) {
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

	if err, s.SortBy = s.owner.StringParam(s.sortByName); err != nil {
		return
	} else if s.SortBy != "" {
		if _, found := s.sortBys[s.SortBy]; !found {
			return fmt.Errorf("'%s' is not valid for %s", s.SortBy, s.sortByName)
		}
	}

	if err, s.Sort = s.owner.StringParam(s.sortName); err != nil {
		return
	} else if s.Sort != "asc" && s.Sort != "desc" {
		return fmt.Errorf("'%s' is not valid for %s", s.Sort, s.sortName)
	}

	if err, s.Limit = s.owner.IntParam(s.limitName); err != nil {
		return
	}

	return
}
