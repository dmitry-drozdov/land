package worker

import (
	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/errgroup"
)

func Iterate[T any](data []T, numThreads int, f func(T) error) error {
	g := errgroup.Group{}
	g.SetLimit(numThreads)
	bar := progressbar.New(len(data))
	for _, t := range data {
		g.Go(func() error {
			defer bar.Add(1)
			return f(t)
		})
	}
	return g.Wait()
}

func IterateMap[K comparable, V any](data map[K]V, numThreads int, f func(K, V) error) error {
	g := errgroup.Group{}
	g.SetLimit(numThreads)
	bar := progressbar.New(len(data))
	for k, v := range data {
		g.Go(func() error {
			defer bar.Add(1)
			return f(k, v)
		})
	}
	return g.Wait()
}
