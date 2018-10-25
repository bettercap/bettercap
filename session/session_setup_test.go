package session

import "testing"

func Test_containsCapitals(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test all alpha lowercase",
			args: args{s: "abcdefghijklmnopqrstuvwxyz"},
			want: false,
		},
		{
			name: "Test all alpha uppercase",
			args: args{s: "ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
			want: true,
		},
		{
			name: "Test special chars",
			args: args{s: "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"},
			want: false,
		},
		// Add test for UTF8 ?
		// {
		// 	name: "Test special UTF-8 chars",
		// 	args: args{s: "€©¶αϚϴЈ"},
		// 	want: false,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsCapitals(tt.args.s); got != tt.want {
				t.Errorf("containsCapitals() = %v, want %v", got, tt.want)
			}
		})
	}
}
