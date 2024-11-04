package client

import (
	"errors"
	"sync"
)

var (
	// ErrNotFound indicates that a request returned no results.
	ErrNotFound = errors.New("no records found for requested object")
)

// import lop "github.com/samber/lo/parallel"

// parallelMap manipulates a slice and transforms it to a slice of another type.
// `iteratee` is call in parallel. Result keep the same order.
func parallelMap[T any, R any](collection []T, iteratee func(T, int) R) []R {
	result := make([]R, len(collection))

	var wg sync.WaitGroup
	wg.Add(len(collection))
	// TODO probably should limit concurrency
	for i, item := range collection {
		go func(_item T, _i int) {
			res := iteratee(_item, _i)

			result[_i] = res

			wg.Done()
		}(item, i)
	}

	wg.Wait()

	return result
}

// genericGet iterates over our different multiClients and invokes our Get Request function for each Type, returning an array of the result type and an error.
func genericGet[T any](mc *MultiClient, get func(client Client) ([]T, error)) ([]T, error) {
	for _, client := range mc.clients {
		result, err := get(client)
		if err == nil {
			// We found it
			return result, nil
		}
		if errors.Is(err, ErrNotFound) {
			// not found, try the next one
			continue
		}
		// TODO advanced: should we ignore errors or collect them until we find the blob.
		return nil, err
	}
	// failed to find it
	return nil, ErrNotFound
}
