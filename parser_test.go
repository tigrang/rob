package main

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := map[string]struct {
		output   string
		expected []outputLine
	}{
		"empty": {
			output:   "",
			expected: nil,
		},
		"single line": {
			output:   "my log output",
			expected: []outputLine{outputLine{Content: "my log output"}},
		},
		"single line with code reference": {
			output: "testdata/sample.go:1:1: fake error",
			expected: []outputLine{
				{
					Content: "testdata/sample.go:1:1: fake error",
					Codeblock: &codeblock{
						LineNum:      1,
						ColNum:       0,
						StartLineNum: 1,
						Code: []codeLine{
							{
								LineNum: 1,
								Content: "package testdata",
							},
						},
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := parse(test.output, ".")
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("actual = %v\nwant %v", actual, test.expected)
			}
		})
	}
}
