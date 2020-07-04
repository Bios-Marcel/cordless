package times

import (
	"encoding/json"
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

func TestMarshalDuration(t *testing.T) {
	duration := Duration(6 * time.Second)
	resultBytes, err := json.Marshal(duration)
	if err != nil {
		t.Errorf("error marshalling duration: %s", err)
	}
	result := string(resultBytes)
	if result != "\"6s\"" {
		t.Errorf("duration should've been '\"6s\"', but was '%s'", result)
	}
}

func TestUnmarshalDuration(t *testing.T) {
	var duration Duration
	err := json.Unmarshal([]byte("\"6s\""), &duration)
	if err != nil {
		t.Errorf("error marshalling duration: %s", err)
	}
	expected := Duration(time.Second * 6)
	if duration != expected {
		t.Errorf("duration should've been '%v', but was '%v'", expected, duration)
	}
}
