package session

import (
	"sync"
	"testing"
	"time"
)

func TestNewEvent(t *testing.T) {

	type args struct {
		tag  string
		data interface{}
	}
	tests := []struct {
		name string
		args args
		want Event
	}{
		{
			name: "Create new event with nil data",
			args: args{"tag", nil},
			want: Event{Tag: "tag", Data: nil},
		},
		{
			name: "Create new event with string data",
			args: args{"tag", "test string"},
			want: Event{Tag: "tag", Data: "test string"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewEvent(tt.args.tag, tt.args.data)
			if got.Data != tt.args.data {
				t.Errorf("Expected %+v data, got %+v", tt.args.data, got.Data)
			}
			if got.Tag != tt.args.tag {
				t.Errorf("Expected %+v Tag, got %+v", tt.args.tag, got.Tag)
			}
		})
	}
}

func TestNewEventPool(t *testing.T) {
	type args struct {
		debug  bool
		silent bool
	}
	tests := []struct {
		name string
		args args
		want *EventPool
	}{
		{
			name: "Test creating debug event pool",
			args: args{true, false},
			want: &EventPool{debug: true, silent: false},
		},
		{
			name: "Test creating silent and event pool",
			args: args{true, false},
			want: &EventPool{debug: true, silent: false},
		},
		// {
		// 	name: "Test creating silent and debug event pool",
		// 	args: args{true, true},
		// 	want: nil,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewEventPool(tt.args.debug, tt.args.silent)
			if got == nil {
				t.Fatal("NewEventPool() returned unexpected nil")
			}
			if got.silent != tt.want.silent {
				t.Errorf("Didn't get correct silent var %v, want %v", got.silent, tt.want.silent)
			}
		})
	}
}

func TestEventPool_SetSilent(t *testing.T) {
	type fields struct {
		Mutex     *sync.Mutex
		debug     bool
		silent    bool
		events    []Event
		listeners []chan Event
	}
	type args struct {
		s bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "Set silent on non-silent event pool",
			fields: fields{silent: false, debug: false, Mutex: &sync.Mutex{}},
			args:   args{s: true},
		},
		{
			name:   "Set silent on silent event pool",
			fields: fields{silent: true, debug: false, Mutex: &sync.Mutex{}},
			args:   args{s: true},
		},
		{
			name:   "Set non-silent on non-silent event pool",
			fields: fields{silent: false, debug: false, Mutex: &sync.Mutex{}},
			args:   args{s: false},
		},
		{
			name:   "Set silent on silent event pool",
			fields: fields{silent: true, debug: false, Mutex: &sync.Mutex{}},
			args:   args{s: false},
		},
		{
			name:   "Set silent on non-silent and debug event pool",
			fields: fields{silent: false, debug: true, Mutex: &sync.Mutex{}},
			args:   args{s: true},
		},
		{
			name:   "Set non-silent on non-silent and debug event pool",
			fields: fields{silent: false, debug: true, Mutex: &sync.Mutex{}},
			args:   args{s: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &EventPool{
				Mutex:     tt.fields.Mutex,
				debug:     tt.fields.debug,
				silent:    tt.fields.silent,
				events:    tt.fields.events,
				listeners: tt.fields.listeners,
			}
			p.SetSilent(tt.args.s)
			if p.silent != tt.args.s {
				t.Error("p.SetSilent didn't set the eventpool to silent")
			}
		})
	}
}

func TestEventPool_SetDebug(t *testing.T) {
	type fields struct {
		Mutex     *sync.Mutex
		debug     bool
		silent    bool
		events    []Event
		listeners []chan Event
	}

	type args struct {
		s bool
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "Set debug on non-debug event pool",
			fields: fields{silent: false, debug: false, Mutex: &sync.Mutex{}},
			args:   args{s: true},
		},
		{
			name:   "Set debug on debug event pool",
			fields: fields{silent: false, debug: true, Mutex: &sync.Mutex{}},
			args:   args{s: true},
		},
		{
			name:   "Set non-debug on non-debug event pool",
			fields: fields{silent: false, debug: false, Mutex: &sync.Mutex{}},
			args:   args{s: false},
		},
		{
			name:   "Set non-debug on debug event pool",
			fields: fields{silent: false, debug: true, Mutex: &sync.Mutex{}},
			args:   args{s: false},
		},
		{
			name:   "Set silent on non-silent and debug event pool",
			fields: fields{silent: false, debug: true, Mutex: &sync.Mutex{}},
			args:   args{s: true},
		},
		{
			name:   "Set non-silent on non-silent and debug event pool",
			fields: fields{silent: false, debug: true, Mutex: &sync.Mutex{}},
			args:   args{s: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &EventPool{
				Mutex:     tt.fields.Mutex,
				debug:     tt.fields.debug,
				silent:    tt.fields.silent,
				events:    tt.fields.events,
				listeners: tt.fields.listeners,
			}
			p.SetDebug(tt.args.s)
			if p.debug != tt.args.s {
				t.Error("p.SetDebug didn't set the eventpool to debug")
			}
		})
	}
}

func TestEventPool_Clear(t *testing.T) {
	type fields struct {
		Mutex     *sync.Mutex
		debug     bool
		silent    bool
		events    []Event
		listeners []chan Event
	}

	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "Clear events on empty list",
			fields: fields{debug: false, silent: false, events: []Event{}, Mutex: &sync.Mutex{}},
		},
		{
			name:   "Clear events",
			fields: fields{debug: false, silent: false, events: []Event{{Tag: "meh", Data: "something", Time: time.Now()}}, Mutex: &sync.Mutex{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &EventPool{
				Mutex:     tt.fields.Mutex,
				debug:     tt.fields.debug,
				silent:    tt.fields.silent,
				events:    tt.fields.events,
				listeners: tt.fields.listeners,
			}
			p.Clear()
			if len(p.events) != 0 {
				t.Errorf("Expected empty list after clear, got %d", len(p.events))
			}
		})
	}
}

func TestEventPool_Add(t *testing.T) {
	type fields struct {
		Mutex     *sync.Mutex
		debug     bool
		silent    bool
		events    []Event
		listeners []chan Event
	}
	type args struct {
		tag  string
		data interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "Add event with nil data on empty event list",
			fields: fields{debug: false, silent: false, events: []Event{}, Mutex: &sync.Mutex{}},
			args:   args{tag: "tag with empty data", data: nil},
		},
		{
			name:   "Add event with nil data",
			fields: fields{debug: false, silent: false, events: []Event{{Tag: "meh", Data: "something", Time: time.Now()}}, Mutex: &sync.Mutex{}},
			args:   args{tag: "tag with empty data", data: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &EventPool{
				Mutex:     tt.fields.Mutex,
				debug:     tt.fields.debug,
				silent:    tt.fields.silent,
				events:    tt.fields.events,
				listeners: tt.fields.listeners,
			}
			eventsList := tt.fields.events[:]
			// It's prepended
			eventsList = append([]Event{{Tag: tt.args.tag, Data: tt.args.data}}, eventsList...)
			p.Add(tt.args.tag, tt.args.data)
			t.Logf("eventsList : %+v", eventsList)
			for index, e := range eventsList {
				if e.Tag != p.events[index].Tag {
					t.Errorf("Tag mismatch, got %s want %s", p.events[index].Tag, e.Tag)
				}
			}
		})
	}
}
