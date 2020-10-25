package version

import "testing"

func Test_isCurrentOlder(t *testing.T) {
	type args struct {
		current string
		other   string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "current older",
			args: args{
				current: "2018-10-10",
				other:   "2018-10-11",
			},
			want: true,
		}, {
			name: "same",
			args: args{
				current: "2018-10-11",
				other:   "2018-10-11",
			},
			want: false,
		}, {
			name: "current newer",
			args: args{
				current: "2018-10-12",
				other:   "2018-10-11",
			},
			want: false,
		}, {
			name: "current day higher, but month and year older",
			args: args{
				current: "2018-09-12",
				other:   "2018-10-11",
			},
			want: true,
		}, {
			name: "current older; additional gibberish behind current",
			args: args{
				current: "2018-10-10R2",
				other:   "2018-10-11",
			},
			want: true,
		}, {
			name: "current older; additional gibberish behind other",
			args: args{
				current: "2018-10-10",
				other:   "2018-10-11R2",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLocalOlderThanRemote(tt.args.current, tt.args.other); got != tt.want {
				t.Errorf("isCurrentOlder() = %v, want %v", got, tt.want)
			}
		})
	}
}
