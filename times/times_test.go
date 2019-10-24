package times

import (
	"testing"
	"time"
)

func TestAreDatesTheSameDay(t *testing.T) {
	type args struct {
		t1 time.Time
		t2 time.Time
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "t1 earlier year",
			args: args{
				t1: time.Date(2000, 10, 10, 10, 10, 10, 0, time.UTC),
				t2: time.Date(2001, 10, 10, 10, 10, 10, 0, time.UTC),
			},
			want: false,
		}, {
			name: "t1 earlier month",
			args: args{
				t1: time.Date(2000, 10, 10, 10, 10, 10, 0, time.UTC),
				t2: time.Date(2000, 11, 10, 10, 10, 10, 0, time.UTC),
			},
			want: false,
		}, {
			name: "t1 earlier day",
			args: args{
				t1: time.Date(2000, 10, 10, 10, 10, 10, 0, time.UTC),
				t2: time.Date(2000, 10, 11, 10, 10, 10, 0, time.UTC),
			},
			want: false,
		}, {
			name: "t1 earlier hour",
			args: args{
				t1: time.Date(2000, 10, 10, 10, 10, 10, 0, time.UTC),
				t2: time.Date(2000, 10, 10, 11, 10, 10, 0, time.UTC),
			},
			want: true,
		}, {
			name: "t1 earlier minute",
			args: args{
				t1: time.Date(2000, 10, 10, 10, 10, 10, 0, time.UTC),
				t2: time.Date(2000, 10, 10, 10, 11, 10, 0, time.UTC),
			},
			want: true,
		}, {
			name: "t1 earlier second",
			args: args{
				t1: time.Date(2000, 10, 10, 10, 10, 10, 0, time.UTC),
				t2: time.Date(2000, 10, 10, 10, 10, 11, 0, time.UTC),
			},
			want: true,
		}, {
			name: "t1 earlier nanosecond",
			args: args{
				t1: time.Date(2000, 10, 10, 10, 10, 10, 0, time.UTC),
				t2: time.Date(2000, 10, 10, 10, 10, 10, 1, time.UTC),
			},
			want: true,
		}, {
			name: "same",
			args: args{
				t1: time.Date(2000, 10, 10, 10, 10, 10, 0, time.UTC),
				t2: time.Date(2000, 10, 10, 10, 10, 10, 0, time.UTC),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AreDatesTheSameDay(tt.args.t1, tt.args.t2); got != tt.want {
				t.Errorf("AreDatesTheSameDay() = %v, want %v", got, tt.want)
			}
		})
	}
}
