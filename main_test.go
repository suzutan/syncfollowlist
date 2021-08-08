package main

import "testing"

func TestInt64ListDivide(t *testing.T) {

	tests := []struct {
		name       string
		mainList   []int64
		divideList []int64
		want       []int64
	}{
		{
			name:       "aT",
			mainList:   []int64{1, 2, 3, 4, 5}, // follows
			divideList: []int64{1, 2, 3, 5, 6}, // list members
			want:       []int64{4},             // add 4 (follows - list members)
		},
		{
			name:       "dT",
			mainList:   []int64{1, 2, 3, 4, 6}, // list members
			divideList: []int64{1, 2, 3, 4, 5}, // follows
			want:       []int64{6},             // remove 6 ( list members - follows)
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Int64ListDivide(tt.mainList, tt.divideList)
			if len(got) != len(tt.want) {
				t.Errorf("Int64ListDivide() = %v, want %v", got, tt.want)
			}

			for k := range got {
				if got[k] != tt.want[k] {
					t.Errorf("Int64ListDivide() = %v, want %v", got, tt.want)
					break
				}
			}
		})
	}
}
