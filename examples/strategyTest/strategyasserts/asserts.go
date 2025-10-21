package strategyasserts

import "embed"

var (
	//go:embed *.txt
	AllStrategy embed.FS
)