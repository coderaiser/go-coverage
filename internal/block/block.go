package block

type Block struct {
	File  string
	Start int
	End   int
	Count int
	Lines []string
	Color bool
}
