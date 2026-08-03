package coverage

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

type helpConfig struct {
	Flags map[string]string `toml:"flags"`
	Env   map[string]string `toml:"env"`
}

func Help() string {
	return HelpFromTOML(helpTOMLBytes)
}

func HelpFromTOML(data []byte) string {
	var cfg helpConfig

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return "usage: coverage [options]\n(help unavailable)"
	}

	var b strings.Builder
	b.WriteString("usage: coverage [options]\n\nflags:\n")

	flagOrder := []string{
		"-f",
		"-r, --report",
		"--code-frame",
		"-v, --version",
		"-h, --help",
	}

	for _, flag := range flagOrder {
		if desc, ok := cfg.Flags[flag]; ok {
			fmt.Fprintf(&b, "  %-22s %s\n", flag, desc)
		}
	}

	if len(cfg.Env) > 0 {
		b.WriteString("\nenvironment variables:\n")
		envOrder := []string{"COVERAGE=codeframe", "COVERAGE=lines"}

		for _, key := range envOrder {
			if desc, ok := cfg.Env[key]; ok {
				fmt.Fprintf(&b, "  %-22s %s\n", key, desc)
			}
		}
	}

	return b.String()
}
