package main

import (
	"sort"
	"testing"
)

func TestAlphabetic(t *testing.T) {
	for _, tc := range []struct{
		name string
		strings alphabetic
		want []string
	} {
		{
			name: "basic",
			strings:alphabetic{"a.b", "d.c", "b.e"},
			want: []string{"a.b", "b.e", "d.c"},
		},
		{
			name: "withKeysAndValues",
			strings:alphabetic{"bKey.subKey=value", "aKey.sub=value", "cKey.subKey=value"},
			want: []string{"aKey.sub=value", "bKey.subKey=value", "cKey.subKey=value"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(tc.strings)
			for idx := range tc.strings {
				if tc.strings[idx] != tc.want[idx] {
					t.Fatalf("incorrect sort\n expected %s\n got %s", tc.strings[idx], tc.want[idx])
				}
			}
		})
	}
}
