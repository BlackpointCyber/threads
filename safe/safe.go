package safe

import "sync"

func Get[T any](mux *sync.Mutex, ref *T) T {
	mux.Lock()
	defer mux.Unlock()

	return *ref
}

func Set[T any](mux *sync.Mutex, ref *T, v T) {
	mux.Lock()
	defer mux.Unlock()

	*ref = v
}
