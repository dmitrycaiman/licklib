package threadsafe

import "sync"

// Semaphore обслуживает квоту на одновременное использование ресурсов.
// Превышение квоты делает Aquire блокирующим до момента, пока она не будет освобождена через Release.
type Semaphore struct {
	cond           *sync.Cond
	quota, counter int
}

func NewSemaphore(quota int) *Semaphore {
	return &Semaphore{quota: quota, cond: sync.NewCond(new(sync.Mutex))}
}

func (slf *Semaphore) Aquire() {
	slf.cond.L.Lock()
	defer slf.cond.L.Unlock()

	for slf.counter >= slf.quota {
		slf.cond.Wait()
	}

	slf.counter++
}

func (slf *Semaphore) Release() {
	slf.cond.L.Lock()
	defer slf.cond.L.Unlock()

	if slf.counter > 0 {
		slf.counter--
	}
	slf.cond.Signal()
}
