package formatter_code_frame

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coderaiser/go-coverage/internal/block"
)

type CodeFrame struct{}

func (CodeFrame) Format(b block.Block) string {
	prefix := ""
	if os.Getenv("TERMINAL_EMULATOR") == "JetBrains-JediTerm" || filepath.IsAbs(b.File) {
		prefix = "file://"
	}

	var header string
	if b.Start == b.End {
		header = fmt.Sprintf("%s%s:%d: %d", prefix, b.File, b.Start, b.Start)
	} else {
		header = fmt.Sprintf("%s%s:%d: %d-%d", prefix, b.File, b.Start, b.Start, b.End)
	}

	if len(b.Lines) == 0 {
		return header
	}

	red, reset, dim := "", "", ""
	if b.Color {
		red = "\033[31m"
		reset = "\033[0m"
		dim = "\033[2m"
	}

	var sb strings.Builder
	sb.WriteString(red + header + reset + "\n\n")

	for i, line := range b.Lines {
		fmt.Fprintf(&sb, "%s%4d%s | %s\n", dim, b.Start+i, reset, line)
	}

	return strings.TrimRight(sb.String(), "\n")
}
