package formatter_json_lines

import (
	"encoding/json"

	"github.com/coderaiser/go-coverage/internal/block"
)

type JSONLines struct{}

type blockJSON struct {
	File  string `json:"file"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

func (JSONLines) Format(b block.Block) string {
	data, _ := json.Marshal(blockJSON{
		File:  b.File,
		Start: b.Start,
		End:   b.End,
	})

	return string(data)
}
