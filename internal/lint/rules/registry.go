package rules

import "github.com/coderaiser/go-coverage/internal/lint/rule"

var All = []rule.Rule{
	&AssertCount{},
	&NoSkip{},
	&NoEqualSlice{},
	&RequireTEnd{},
	&ExtractResultFromAssertion{},
}
