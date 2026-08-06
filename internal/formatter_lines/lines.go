package formatter_lines

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/coderaiser/go-coverage/internal/block"
)

type Lines struct{}

func (Lines) Format(b block.Block) string {
	prefix := ""

	if os.Getenv("TERMINAL_EMULATOR") == "JetBrains-JediTerm" || filepath.IsAbs(b.File) {
		prefix = "file://"
	}

	if b.Start == b.End {
		return fmt.Sprintf("%s%s:%d: %d", prefix, b.File, b.Start, b.Start)
	}

	return fmt.Sprintf("%s%s:%d: %d-%d", prefix, b.File, b.Start, b.Start, b.End)
}
