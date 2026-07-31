package block

type Block struct {
	File  string
	Start int
	End   int
	Lines []string
	Color bool
}
