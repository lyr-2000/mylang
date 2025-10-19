module github.com/lyr-2000/mylang

go 1.25.1

require (
	github.com/MetalBlueberry/go-plotly v0.0.0-00010101000000-000000000000
	github.com/spf13/cast v1.10.0
)

require (
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	golang.org/x/sys v0.23.0 // indirect
)

replace github.com/MetalBlueberry/go-plotly => ./pkg/extensions/tradingcharts/go-plotly
