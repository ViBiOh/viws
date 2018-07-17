package viws

import (
	"fmt"
	"testing"
)

func Test_Flags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
		wantType  string
	}{
		{
			`should add string directory param to flags`,
			`directory`,
			`*string`,
		},
		{
			`should add string headers param to flags`,
			`headers`,
			`*string`,
		},
		{
			`should add bool notFound param to flags`,
			`notFound`,
			`*bool`,
		},
		{
			`should add bool spa param to flags`,
			`spa`,
			`*bool`,
		},
		{
			`should add string push param to flags`,
			`push`,
			`*string`,
		},
	}

	for _, testCase := range cases {
		result := Flags(testCase.intention)[testCase.want]

		if result == nil {
			t.Errorf("%s\nFlags() = %+v, want `%s`", testCase.intention, result, testCase.want)
		}

		if fmt.Sprintf(`%T`, result) != testCase.wantType {
			t.Errorf("%s\nFlags() = `%T`, want `%s`", testCase.intention, result, testCase.wantType)
		}
	}
}
