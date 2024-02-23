module gotest

go 1.18

require (
	github.com/segmentio/encoding v0.4.0
	golang.org/x/sync v0.6.0
	golang.org/x/sys v0.0.0-20211110154304-99a53858aa08
	utils v0.0.0
)

require (
	github.com/segmentio/asm v1.1.3 // indirect
	golang.org/x/exp v0.0.0-20240222234643-814bf88cf225 // indirect
)

replace utils => ../utils
