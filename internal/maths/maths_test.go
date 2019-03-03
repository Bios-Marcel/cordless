package maths

import (
	"testing"
)

func TestMax(t *testing.T) {
	type args struct {
		a int
		b int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "equal numbers above 0",
			args: args{1, 1},
			want: 1,
		}, {
			name: "two zeroes",
			args: args{0, 0},
			want: 0,
		}, {
			name: "equal numbers below zero",
			args: args{-1, -1},
			want: -1,
		}, {
			name: "first is bigger",
			args: args{3, 2},
			want: 3,
		}, {
			name: "first is smaller",
			args: args{2, 3},
			want: 3,
		}, {
			name: "both negative and first is bigger",
			args: args{-2, -3},
			want: -2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Max(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("Max() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMin(t *testing.T) {
	type args struct {
		a int
		b int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "equal numbers above 0",
			args: args{1, 1},
			want: 1,
		}, {
			name: "two zeroes",
			args: args{0, 0},
			want: 0,
		}, {
			name: "equal numbers below zero",
			args: args{-1, -1},
			want: -1,
		}, {
			name: "first is bigger",
			args: args{3, 2},
			want: 2,
		}, {
			name: "first is smaller",
			args: args{2, 3},
			want: 2,
		}, {
			name: "both negative and first is bigger",
			args: args{-2, -3},
			want: -3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Min(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("Min() = %v, want %v", got, tt.want)
			}
		})
	}
}
