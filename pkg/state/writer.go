package state

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type PrismStateWriter struct {
	mu   sync.Mutex
	mmap *MappedFile
	ptr  *PrismRuntimeState
}

func NewPrismStateWriter(path string) (*PrismStateWriter, error) {
	mmap, err := CreateMappedFile(path, PrismRuntimeStateSize)
	if err != nil {
		return nil, err
	}

	ptr := (*PrismRuntimeState)(mmap.AsPtr())
	atomic.StoreUint64(&ptr.Version, 0)

	return &PrismStateWriter{
		mmap: mmap,
		ptr:  ptr,
	}, nil
}

func (w *PrismStateWriter) SetInstance(name string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.beginWrite()
	w.ptr.SetInstance(name)
	w.endWrite()
}

func (w *PrismStateWriter) Update(fn func(*PrismRuntimeState)) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.beginWrite()
	fn(w.ptr)
	w.endWrite()
}

func (w *PrismStateWriter) SetPrism(index int, name string, pid int32, state PrismEntryState, restarts uint8, startMs int64) error {
	if index < 0 || index >= MaxPrisms {
		return fmt.Errorf("prism index out of range: %d", index)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.beginWrite()

	entry := &w.ptr.Prisms[index]
	entry.SetName(name)
	entry.PID = pid
	entry.State = uint8(state)
	entry.Restarts = restarts
	entry.StartMs = startMs

	if index >= int(w.ptr.PrismCount) {
		w.ptr.PrismCount = uint8(index + 1)
	}

	w.endWrite()
	return nil
}

func (w *PrismStateWriter) RemovePrism(name string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.beginWrite()

	found := -1
	for i := 0; i < int(w.ptr.PrismCount); i++ {
		if w.ptr.Prisms[i].GetName() == name {
			found = i
			break
		}
	}

	if found >= 0 {
		for i := found; i < int(w.ptr.PrismCount)-1; i++ {
			w.ptr.Prisms[i] = w.ptr.Prisms[i+1]
		}
		w.ptr.Prisms[w.ptr.PrismCount-1] = PrismEntry{}
		w.ptr.PrismCount--
	}

	w.endWrite()
}

func (w *PrismStateWriter) SetForeground(name string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.beginWrite()

	w.ptr.SetFgPrism(name)

	for i := 0; i < int(w.ptr.PrismCount); i++ {
		if w.ptr.Prisms[i].GetName() == name {
			w.ptr.Prisms[i].State = uint8(PrismStateFg)
		} else {
			w.ptr.Prisms[i].State = uint8(PrismStateBg)
		}
	}

	w.endWrite()
}

func (w *PrismStateWriter) AddPrism(name string, pid int32, fg bool) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.ptr.PrismCount >= MaxPrisms {
		return -1, fmt.Errorf("max prisms reached: %d", MaxPrisms)
	}

	w.beginWrite()

	idx := int(w.ptr.PrismCount)
	entry := &w.ptr.Prisms[idx]
	entry.SetName(name)
	entry.PID = pid
	if fg {
		entry.State = uint8(PrismStateFg)
		w.ptr.SetFgPrism(name)
		for i := 0; i < idx; i++ {
			w.ptr.Prisms[i].State = uint8(PrismStateBg)
		}
	} else {
		entry.State = uint8(PrismStateBg)
	}
	entry.Restarts = 0
	entry.StartMs = time.Now().UnixMilli()
	w.ptr.PrismCount++

	w.endWrite()
	return idx, nil
}

func (w *PrismStateWriter) beginWrite() {
	v := atomic.LoadUint64(&w.ptr.Version)
	atomic.StoreUint64(&w.ptr.Version, v+1)
}

func (w *PrismStateWriter) endWrite() {
	v := atomic.LoadUint64(&w.ptr.Version)
	atomic.StoreUint64(&w.ptr.Version, v+1)
	w.mmap.Sync()
}

func (w *PrismStateWriter) Sync() error {
	return w.mmap.Sync()
}

func (w *PrismStateWriter) Close() error {
	return w.mmap.Close()
}

func (w *PrismStateWriter) Remove() error {
	return w.mmap.Remove()
}

func (w *PrismStateWriter) Path() string {
	return w.mmap.Path()
}

type ShinedStateWriter struct {
	mu   sync.Mutex
	mmap *MappedFile
	ptr  *ShinedState
}

func NewShinedStateWriter(path string) (*ShinedStateWriter, error) {
	mmap, err := CreateMappedFile(path, ShinedStateSize)
	if err != nil {
		return nil, err
	}

	ptr := (*ShinedState)(unsafe.Pointer(&mmap.Data()[0]))
	atomic.StoreUint64(&ptr.Version, 0)

	return &ShinedStateWriter{
		mmap: mmap,
		ptr:  ptr,
	}, nil
}

func (w *ShinedStateWriter) Update(fn func(*ShinedState)) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.beginWrite()
	fn(w.ptr)
	w.endWrite()
}

func (w *ShinedStateWriter) AddPanel(instance, name string, pid int32, healthy bool) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.ptr.PanelCount >= MaxPanels {
		return -1, fmt.Errorf("max panels reached: %d", MaxPanels)
	}

	w.beginWrite()

	idx := int(w.ptr.PanelCount)
	entry := &w.ptr.Panels[idx]
	entry.SetInstance(instance)
	entry.SetName(name)
	entry.PID = pid
	if healthy {
		entry.Healthy = 1
	} else {
		entry.Healthy = 0
	}
	w.ptr.PanelCount++

	w.endWrite()
	return idx, nil
}

func (w *ShinedStateWriter) RemovePanel(instance string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.beginWrite()

	found := -1
	for i := 0; i < int(w.ptr.PanelCount); i++ {
		if w.ptr.Panels[i].GetInstance() == instance {
			found = i
			break
		}
	}

	if found >= 0 {
		for i := found; i < int(w.ptr.PanelCount)-1; i++ {
			w.ptr.Panels[i] = w.ptr.Panels[i+1]
		}
		w.ptr.Panels[w.ptr.PanelCount-1] = PanelEntry{}
		w.ptr.PanelCount--
	}

	w.endWrite()
}

func (w *ShinedStateWriter) SetPanelHealth(instance string, healthy bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.beginWrite()

	for i := 0; i < int(w.ptr.PanelCount); i++ {
		if w.ptr.Panels[i].GetInstance() == instance {
			if healthy {
				w.ptr.Panels[i].Healthy = 1
			} else {
				w.ptr.Panels[i].Healthy = 0
			}
			break
		}
	}

	w.endWrite()
}

func (w *ShinedStateWriter) beginWrite() {
	v := atomic.LoadUint64(&w.ptr.Version)
	atomic.StoreUint64(&w.ptr.Version, v+1)
}

func (w *ShinedStateWriter) endWrite() {
	v := atomic.LoadUint64(&w.ptr.Version)
	atomic.StoreUint64(&w.ptr.Version, v+1)
	w.mmap.Sync()
}

func (w *ShinedStateWriter) Sync() error {
	return w.mmap.Sync()
}

func (w *ShinedStateWriter) Close() error {
	return w.mmap.Close()
}

func (w *ShinedStateWriter) Remove() error {
	return w.mmap.Remove()
}

func (w *ShinedStateWriter) Path() string {
	return w.mmap.Path()
}
