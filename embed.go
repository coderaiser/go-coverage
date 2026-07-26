package coverage

import _ "embed"

//go:embed package.json
var packageJSONBytes []byte

//go:embed help.toml
var helpTOMLBytes []byte
