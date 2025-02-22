module graphqltest

go 1.22.5

require github.com/graph-gophers/graphql-go v1.5.0

require (
	github.com/segmentio/asm v1.1.3 // indirect
	golang.org/x/sys v0.25.0 // indirect
)

require (
	github.com/segmentio/encoding v0.4.0
	golang.org/x/sync v0.8.0
	utils v0.0.0
)

replace utils => ../utils
