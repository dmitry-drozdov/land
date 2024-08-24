module gotest

go 1.22.5

require (
	github.com/segmentio/encoding v0.4.0
	golang.org/x/sync v0.7.0
	golang.org/x/sys v0.22.0
	utils v0.0.0
)

require (
	github.com/mohae/shuffle v0.0.0-20160809015857-b0f723480796
	github.com/otiai10/copy v1.14.0
	github.com/segmentio/asm v1.1.3 // indirect
	golang.org/x/exp v0.0.0-20240222234643-814bf88cf225
)

replace utils => ../utils
