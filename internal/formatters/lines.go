package formatters

import (
	"fmt"

	"coderaiser/go-coverage/internal/block"
)

type Lines struct{}

func (Lines) Format(b block.Block) string {
	if b.Start == b.End {
		return fmt.Sprintf("file://%s:%d: %d", b.File, b.Start, b.Start)
	}

	return fmt.Sprintf("file://%s:%d: %d-%d", b.File, b.Start, b.Start, b.End)
}
