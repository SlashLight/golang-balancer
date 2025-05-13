package balancer

import (
	"container/heap"
	"fmt"
	"net/http"
	"sync"
)

type BackendConnections struct {
	connections uint64
	Back        *Backend
}

type BackendHeap []*BackendConnections

func (h BackendHeap) Len() int           { return len(h) }
func (h BackendHeap) Less(i, j int) bool { return h[i].connections < h[j].connections }

func (h BackendHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Back.Index = i
	h[j].Back.Index = j
}

func (h *BackendHeap) Push(x interface{}) {
	backend := x.(*BackendConnections)
	backend.Back.Index = len(*h)
	*h = append(*h, backend)
}

func (h *BackendHeap) Pop() interface{} {
	old := *h
	n := len(old)
	backend := old[n-1]
	backend.Back.Index = -1
	*h = old[0 : n-1]
	return backend
}

type LeastConnectionsBalancer struct {
	backends BackendHeap
	mu       sync.RWMutex
}

func NewLeastConnectionBalancer(backendsURLs []string) (*LeastConnectionsBalancer, error) {
	backends, err := GetBackendsFromURLS(backendsURLs)
	if err != nil {
		return nil, fmt.Errorf("Error at creating Least-connections balancer: %w", err)
	}
	backendWithCons := make(BackendHeap, len(backendsURLs))
	for idx, back := range backends {
		backendWithCons[idx] = &BackendConnections{
			connections: 0,
			Back:        back,
		}
	}

	heap.Init(&backendWithCons)

	return &LeastConnectionsBalancer{
		backends: backendWithCons,
		mu:       sync.RWMutex{},
	}, nil
}

func (lc *LeastConnectionsBalancer) Next(r *http.Request) (*Backend, error) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	backend := lc.backends[0]
	backend.connections++
	heap.Fix(&lc.backends, 0)

	return backend.Back, nil
}

func (lc *LeastConnectionsBalancer) Release(backend *BackendConnections) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	backend.connections--
	heap.Fix(&lc.backends, backend.Back.Index)
}

func (lc *LeastConnectionsBalancer) AddNewBackend(back *Backend) {
	heap.Push(&lc.backends, back)
}

func (lc *LeastConnectionsBalancer) RemoveBackend(idx int) {
	heap.Remove(&lc.backends, idx)
}
