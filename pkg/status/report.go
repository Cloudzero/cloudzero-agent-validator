package status

import (
	"sync"
)

var _ interface {
	Accessor
} = (*builder)(nil) // integrity check

// Accessor allows for low-level access to the Report
type Accessor interface {
	AddCheck(...*StatusCheck)
	WriteToReport(func(*ClusterStatus))
	ReadFromReport(func(*ClusterStatus))
}

type builder struct {
	report  *ClusterStatus
	lock    *sync.RWMutex
	onWrite []func(*ClusterStatus)
}

func NewAccessor(s *ClusterStatus, onWrite ...func(*ClusterStatus)) Accessor {
	return &builder{
		report:  s,
		lock:    &sync.RWMutex{},
		onWrite: onWrite,
	}
}

func (b builder) onWriteEvent() {
	for _, fn := range b.onWrite {
		fn(b.report)
	}
}

func (b builder) WriteToReport(fn func(*ClusterStatus)) {
	b.lock.Lock()
	defer b.lock.Unlock()

	fn(b.report)
	b.onWriteEvent()
}

func (b builder) ReadFromReport(fn func(*ClusterStatus)) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	fn(b.report)
}

func (b builder) AddCheck(c ...*StatusCheck) {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.report.Checks = append(b.report.Checks, c...)
	b.onWriteEvent()
}
