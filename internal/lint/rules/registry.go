package rules

import "coderaiser/go-coverage/internal/lint/rule"

var All = []rule.Rule{
	&AssertCount{},
	&NoSkip{},
	&NoEqualSlice{},
	&RequireTEnd{},
}
