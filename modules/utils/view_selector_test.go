package utils

import (
	"regexp"
	"sync"
	"testing"

	"github.com/bettercap/bettercap/v2/session"
)

var (
	testSession *session.Session
	sessionOnce sync.Once
)

func createMockSession(t *testing.T) *session.Session {
	sessionOnce.Do(func() {
		var err error
		testSession, err = session.New()
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
	})
	return testSession
}

type mockModule struct {
	session.SessionModule
}

func newMockModule(s *session.Session) *mockModule {
	return &mockModule{
		SessionModule: session.NewSessionModule("test", s),
	}
}

func TestViewSelectorFor(t *testing.T) {
	s := createMockSession(t)
	m := newMockModule(s)

	sortFields := []string{"name", "mac", "seen"}
	defExpression := "seen desc"
	prefix := "test"

	vs := ViewSelectorFor(&m.SessionModule, prefix, sortFields, defExpression)

	if vs == nil {
		t.Fatal("ViewSelectorFor returned nil")
	}

	if vs.owner != &m.SessionModule {
		t.Error("ViewSelector owner not set correctly")
	}

	if vs.filterName != "test.filter" {
		t.Errorf("filterName = %s, want test.filter", vs.filterName)
	}

	if vs.sortName != "test.sort" {
		t.Errorf("sortName = %s, want test.sort", vs.sortName)
	}

	if vs.limitName != "test.limit" {
		t.Errorf("limitName = %s, want test.limit", vs.limitName)
	}

	// Check that parameters were added by trying to retrieve them
	if err, _ := m.SessionModule.StringParam("test.filter"); err != nil {
		t.Error("filter parameter not accessible")
	}
	if err, _ := m.SessionModule.StringParam("test.sort"); err != nil {
		t.Error("sort parameter not accessible")
	}
	if err, _ := m.SessionModule.IntParam("test.limit"); err != nil {
		t.Error("limit parameter not accessible")
	}

	// Check default sorting
	if vs.SortField != "seen" {
		t.Errorf("Default SortField = %s, want seen", vs.SortField)
	}
	if vs.Sort != "desc" {
		t.Errorf("Default Sort = %s, want desc", vs.Sort)
	}
}

func TestParseFilter(t *testing.T) {
	s := createMockSession(t)
	m := newMockModule(s)
	vs := ViewSelectorFor(&m.SessionModule, "test", []string{"name"}, "name asc")

	tests := []struct {
		name     string
		filter   string
		wantErr  bool
		wantExpr bool
	}{
		{
			name:     "empty filter",
			filter:   "",
			wantErr:  false,
			wantExpr: false,
		},
		{
			name:     "valid regex",
			filter:   "^test.*",
			wantErr:  false,
			wantExpr: true,
		},
		{
			name:     "invalid regex",
			filter:   "[invalid",
			wantErr:  true,
			wantExpr: false,
		},
		{
			name:     "simple string",
			filter:   "test",
			wantErr:  false,
			wantExpr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the filter parameter
			m.Session.Env.Set("test.filter", tt.filter)

			err := vs.parseFilter()
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFilter() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantExpr && vs.Expression == nil {
				t.Error("Expected Expression to be set, but it's nil")
			}
			if !tt.wantExpr && vs.Expression != nil {
				t.Error("Expected Expression to be nil, but it's set")
			}

			if tt.filter != "" && !tt.wantErr {
				if vs.Filter != tt.filter {
					t.Errorf("Filter = %s, want %s", vs.Filter, tt.filter)
				}
			}
		})
	}
}

func TestParseSorting(t *testing.T) {
	s := createMockSession(t)
	m := newMockModule(s)
	vs := ViewSelectorFor(&m.SessionModule, "test", []string{"name", "mac", "seen"}, "name asc")

	tests := []struct {
		name          string
		sortExpr      string
		wantErr       bool
		wantField     string
		wantDirection string
		wantSymbol    string
	}{
		{
			name:          "name ascending",
			sortExpr:      "name asc",
			wantErr:       false,
			wantField:     "name",
			wantDirection: "asc",
			wantSymbol:    "▴", // Will be colored blue
		},
		{
			name:          "mac descending",
			sortExpr:      "mac desc",
			wantErr:       false,
			wantField:     "mac",
			wantDirection: "desc",
			wantSymbol:    "▾", // Will be colored blue
		},
		{
			name:          "seen descending",
			sortExpr:      "seen desc",
			wantErr:       false,
			wantField:     "seen",
			wantDirection: "desc",
			wantSymbol:    "▾",
		},
		{
			name:          "invalid field",
			sortExpr:      "invalid desc",
			wantErr:       true,
			wantField:     "",
			wantDirection: "",
		},
		{
			name:          "invalid direction",
			sortExpr:      "name invalid",
			wantErr:       true,
			wantField:     "",
			wantDirection: "",
		},
		{
			name:          "malformed expression",
			sortExpr:      "nameDesc",
			wantErr:       true,
			wantField:     "",
			wantDirection: "",
		},
		{
			name:          "empty expression",
			sortExpr:      "",
			wantErr:       true,
			wantField:     "",
			wantDirection: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the sort parameter
			m.Session.Env.Set("test.sort", tt.sortExpr)

			err := vs.parseSorting()
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSorting() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if vs.SortField != tt.wantField {
					t.Errorf("SortField = %s, want %s", vs.SortField, tt.wantField)
				}
				if vs.Sort != tt.wantDirection {
					t.Errorf("Sort = %s, want %s", vs.Sort, tt.wantDirection)
				}
				// Check symbol contains expected character (stripping color codes)
				if !containsSymbol(vs.SortSymbol, tt.wantSymbol) {
					t.Errorf("SortSymbol doesn't contain %s", tt.wantSymbol)
				}
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	s := createMockSession(t)
	m := newMockModule(s)
	vs := ViewSelectorFor(&m.SessionModule, "test", []string{"name", "mac"}, "name asc")

	tests := []struct {
		name      string
		filter    string
		sort      string
		limit     string
		wantErr   bool
		wantLimit int
	}{
		{
			name:      "all valid",
			filter:    "test.*",
			sort:      "mac desc",
			limit:     "10",
			wantErr:   false,
			wantLimit: 10,
		},
		{
			name:      "invalid filter",
			filter:    "[invalid",
			sort:      "name asc",
			limit:     "5",
			wantErr:   true,
			wantLimit: 0,
		},
		{
			name:      "invalid sort",
			filter:    "valid",
			sort:      "invalid field",
			limit:     "5",
			wantErr:   true,
			wantLimit: 0,
		},
		{
			name:      "invalid limit",
			filter:    "valid",
			sort:      "name asc",
			limit:     "not a number",
			wantErr:   true,
			wantLimit: 0,
		},
		{
			name:      "zero limit",
			filter:    "",
			sort:      "name asc",
			limit:     "0",
			wantErr:   false,
			wantLimit: 0,
		},
		{
			name:      "negative limit",
			filter:    "",
			sort:      "name asc",
			limit:     "-1",
			wantErr:   false,
			wantLimit: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set parameters
			m.Session.Env.Set("test.filter", tt.filter)
			m.Session.Env.Set("test.sort", tt.sort)
			m.Session.Env.Set("test.limit", tt.limit)

			err := vs.Update()
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if vs.Limit != tt.wantLimit {
					t.Errorf("Limit = %d, want %d", vs.Limit, tt.wantLimit)
				}
			}
		})
	}
}

func TestFilterCaching(t *testing.T) {
	s := createMockSession(t)
	m := newMockModule(s)
	vs := ViewSelectorFor(&m.SessionModule, "test", []string{"name"}, "name asc")

	// Set initial filter
	m.Session.Env.Set("test.filter", "test1")
	if err := vs.parseFilter(); err != nil {
		t.Fatalf("Failed to parse initial filter: %v", err)
	}

	firstExpr := vs.Expression
	if firstExpr == nil {
		t.Fatal("Expression should not be nil")
	}

	// Parse again with same filter - should use cached expression
	if err := vs.parseFilter(); err != nil {
		t.Fatalf("Failed to parse filter second time: %v", err)
	}

	// The filterPrev mechanism should prevent recompilation
	if vs.filterPrev != "test1" {
		t.Errorf("filterPrev = %s, want test1", vs.filterPrev)
	}

	// Change filter
	m.Session.Env.Set("test.filter", "test2")
	if err := vs.parseFilter(); err != nil {
		t.Fatalf("Failed to parse new filter: %v", err)
	}

	if vs.Filter != "test2" {
		t.Errorf("Filter = %s, want test2", vs.Filter)
	}
	if vs.filterPrev != "test2" {
		t.Errorf("filterPrev = %s, want test2", vs.filterPrev)
	}
}

func TestSortParserRegex(t *testing.T) {
	s := createMockSession(t)
	m := newMockModule(s)

	sortFields := []string{"field1", "field2", "complex_field"}
	vs := ViewSelectorFor(&m.SessionModule, "test", sortFields, "field1 asc")

	// Test the generated regex pattern
	expectedPattern := "(field1|field2|complex_field) (desc|asc)"
	if vs.sortParser != expectedPattern {
		t.Errorf("sortParser = %s, want %s", vs.sortParser, expectedPattern)
	}

	// Test regex compilation
	if vs.sortParse == nil {
		t.Fatal("sortParse regex is nil")
	}

	// Test regex matching
	testCases := []struct {
		expr    string
		matches bool
	}{
		{"field1 asc", true},
		{"field2 desc", true},
		{"complex_field asc", true},
		{"invalid_field asc", false},
		{"field1 invalid", false},
		{"field1asc", false},
		{"", false},
	}

	for _, tc := range testCases {
		matches := vs.sortParse.MatchString(tc.expr)
		if matches != tc.matches {
			t.Errorf("sortParse.MatchString(%q) = %v, want %v", tc.expr, matches, tc.matches)
		}
	}
}

// Helper function to check if a string contains a symbol (ignoring ANSI color codes)
func containsSymbol(s, symbol string) bool {
	// Remove ANSI color codes
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	cleaned := re.ReplaceAllString(s, "")
	return cleaned == symbol
}

// Benchmark tests
func BenchmarkParseFilter(b *testing.B) {
	s, _ := session.New()
	m := newMockModule(s)
	vs := ViewSelectorFor(&m.SessionModule, "test", []string{"name"}, "name asc")

	m.Session.Env.Set("test.filter", "test.*")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vs.parseFilter()
	}
}

func BenchmarkParseSorting(b *testing.B) {
	s, _ := session.New()
	m := newMockModule(s)
	vs := ViewSelectorFor(&m.SessionModule, "test", []string{"name", "mac", "seen"}, "name asc")

	m.Session.Env.Set("test.sort", "mac desc")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vs.parseSorting()
	}
}

func BenchmarkUpdate(b *testing.B) {
	s, _ := session.New()
	m := newMockModule(s)
	vs := ViewSelectorFor(&m.SessionModule, "test", []string{"name", "mac"}, "name asc")

	m.Session.Env.Set("test.filter", "test")
	m.Session.Env.Set("test.sort", "mac desc")
	m.Session.Env.Set("test.limit", "10")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vs.Update()
	}
}
