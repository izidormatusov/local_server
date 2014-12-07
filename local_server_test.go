package main

import (
	"testing"
	"strings"
)

func testEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

var aliasestests = []struct {
	content string
	aliases []string
	output []string
}{
	{"", []string{}, []string{}},
	{
		"127.0.42.42	localhost local",
		[]string{"localhost", "local"},
		[]string{"localhost", "local"},
	},
	{"127.0.42.42 machine", []string{"localhost"}, []string{}},
	{"#127.0.42.42 localhost", []string{"localhost"}, []string{}},
	// Can't be associated with a non-local IP
	{"8.8.8.8 machine", []string{"machine"}, nil},
	{"127.0.42.42 machine", []string{"machine"}, []string{"machine"}},
}

func TestFindAliases(t *testing.T) {
	for _, tt := range aliasestests {
		r := strings.NewReader(tt.content)
		aliases, err := findAliases(r, tt.aliases)

		if tt.output == nil {
			if err == nil {
				t.Errorf("Content %q and aliases %q should cause error",
					tt.content, tt.aliases)
			}
			continue
		}

		if err != nil {
			t.Errorf("Content %q and aliases %q caused error %q",
				tt.content, tt.aliases, err)
		}
		if !testEq(aliases, tt.output) {
			t.Errorf("Content %q and aliases %q => %q, want %q",
				tt.content, tt.aliases, aliases, tt.output)
		}
	}
}
