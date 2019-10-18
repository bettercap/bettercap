package session

import (
	"reflect"
	"regexp"
	"testing"
)

func TestNewModuleHandler(t *testing.T) {
	type args struct {
		name string
		expr string
		desc string
		exec func(args []string) error
	}
	tests := []struct {
		name string
		args args
		want ModuleHandler
	}{
		{
			name: "Test NewModuleHandler empty expr",
			args: args{name: "test name", desc: "test description"},
			want: ModuleHandler{Name: "test name", Description: "test description"},
		},
		{
			name: "Test NewModuleHandler normal",
			args: args{name: "test name", desc: "test description", expr: `[a-z]`},
			want: ModuleHandler{Name: "test name", Description: "test description", Parser: regexp.MustCompile(`[a-z]`)},
		},
		// this test shoud handle panic on bad regexp ?
		// {
		// 	name: "Test NewModuleHandler bad regex expr",
		// 	args: args{name: "test name", desc: "test description", expr: "/abcd.(]"},
		// 	want: ModuleHandler{Name: "test name", Description: "test description"},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModuleHandler(tt.args.name, tt.args.expr, tt.args.desc, tt.args.exec)
			if m.Parser != nil {
				if tt.args.expr == "" {
					t.Errorf("If no regexp were provided, should got nil parser but got %+v", m.Parser)
				}
				if m.Parser.String() != tt.want.Parser.String() {
					t.Errorf("Wrong parser, got %+v want %+v", m.Parser.String(), tt.want.Parser.String())
				}
			}
		})
	}
}

func TestModuleHandler_Help(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		Parser      *regexp.Regexp
		Exec        func(args []string) error
	}
	type args struct {
		padding int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "Test help with no padding",
			fields: fields{Name: "no-padding", Description: "Test without padding"},
			args:   args{padding: 0},
			want:   "  \033[1mno-padding\033[0m : Test without padding\n",
		},
		{
			name:   "Test help with non-needed padding",
			fields: fields{Name: "non-needed padding", Description: "Test with non needed padding (5)"},
			args:   args{padding: 5},
			want:   "  \033[1mnon-needed padding\033[0m : Test with non needed padding (5)\n",
		},
		{
			name:   "Test help with 20 padding",
			fields: fields{Name: "padding", Description: "Test with 20 padding"},
			args:   args{padding: 20},
			want:   "  \033[1m             padding\033[0m : Test with 20 padding\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &ModuleHandler{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Parser:      tt.fields.Parser,
				exec:        tt.fields.Exec,
			}
			if got := h.Help(tt.args.padding); got != tt.want {
				t.Errorf("ModuleHandler.Help() = \n%v, want\n%v", got, tt.want)
			}
		})
	}
}

func TestModuleHandler_Parse(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		Parser      *regexp.Regexp
		Exec        func(args []string) error
	}
	type args struct {
		line string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
		want1  []string
	}{
		{
			name:   "check parse on nil parser match name",
			fields: fields{Name: "parser", Description: "description of the parser", Parser: nil},
			args:   args{line: "parser"},
			want:   true,
			want1:  nil,
		},
		{
			name:   "check parse on nil parser no match name",
			fields: fields{Name: "parser", Description: "description of the parser", Parser: nil},
			args:   args{line: "wrongname"},
			want:   false,
			want1:  nil,
		},
		{
			name:   "check parse on existing parser",
			fields: fields{Name: "parser", Description: "description of the parser", Parser: regexp.MustCompile("(abcd)")},
			args:   args{line: "lol abcd lol"},
			want:   true,
			want1:  []string{"abcd"},
		},
		{
			name:   "check parse on existing parser",
			fields: fields{Name: "parser", Description: "description of the parser", Parser: regexp.MustCompile("(abcd)")},
			args:   args{line: "no match"},
			want:   false,
			want1:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &ModuleHandler{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Parser:      tt.fields.Parser,
				exec:        tt.fields.Exec,
			}
			got, got1 := h.Parse(tt.args.line)
			if got != tt.want {
				t.Errorf("ModuleHandler.Parse() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ModuleHandler.Parse() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestModuleHandler_MarshalJSON(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		Parser      *regexp.Regexp
		Exec        func(args []string) error
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name:    "generating JSON with empty parser",
			fields:  fields{Name: "test name", Description: "test description", Parser: nil},
			want:    []byte(`{"name":"test name","description":"test description","parser":""}`),
			wantErr: false,
		},
		{
			name:    "generating JSON with parser",
			fields:  fields{Name: "test name", Description: "test description", Parser: regexp.MustCompile("abcd")},
			want:    []byte(`{"name":"test name","description":"test description","parser":"abcd"}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := ModuleHandler{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Parser:      tt.fields.Parser,
				exec:        tt.fields.Exec,
			}
			got, err := h.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("ModuleHandler.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != string(tt.want) {
				t.Errorf("ModuleHandler.MarshalJSON() = \n%+v, want \n%+v", string(got), string(tt.want))
			}
		})
	}
}
