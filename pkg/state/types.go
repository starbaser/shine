package state

import (
	"fmt"
	"time"
	"unsafe"
)

const (
	PrismEntrySize        = 80   // bytes per prism entry
	MaxPrisms             = 16   // max prisms per prismctl instance
	PrismRuntimeStateSize = 1424 // total size of PrismRuntimeState (includes alignment padding)

	PanelEntrySize       = 136  // bytes per panel entry
	MaxPanels            = 32   // max panels
	ShinedStateSize    = 4368 // total size of ShinedState

	NameMaxLen = 63 // max length for names (64 bytes with length prefix)
)

type PrismEntryState uint8

const (
	PrismStateBg PrismEntryState = 0 // background
	PrismStateFg PrismEntryState = 1 // foreground
)

func (s PrismEntryState) String() string {
	switch s {
	case PrismStateBg:
		return "bg"
	case PrismStateFg:
		return "fg"
	default:
		return "unknown"
	}
}

type PrismEntry struct {
	NameLen  uint8     // 1 byte: length of name
	Name     [63]byte  // 63 bytes: name (null-padded)
	PID      int32     // 4 bytes: process ID
	State    uint8     // 1 byte: 0=bg, 1=fg
	Restarts uint8     // 1 byte: restart count (capped at 255)
	_padding [2]byte   // 2 bytes: padding for alignment
	StartMs  int64     // 8 bytes: unix ms when started
}

func (e *PrismEntry) GetName() string {
	return string(e.Name[:e.NameLen])
}

func (e *PrismEntry) SetName(name string) {
	if len(name) > NameMaxLen {
		name = name[:NameMaxLen]
	}
	e.NameLen = uint8(len(name))
	copy(e.Name[:], name)
	for i := int(e.NameLen); i < NameMaxLen; i++ {
		e.Name[i] = 0
	}
}

func (e *PrismEntry) GetState() PrismEntryState {
	return PrismEntryState(e.State)
}

func (e *PrismEntry) Uptime() time.Duration {
	if e.StartMs == 0 {
		return 0
	}
	return time.Duration(time.Now().UnixMilli()-e.StartMs) * time.Millisecond
}

func (e *PrismEntry) IsActive() bool {
	return e.PID != 0
}

type PrismRuntimeState struct {
	Version     uint64           // 8 bytes: sequence counter (odd=writing, even=complete)
	InstanceLen uint8            // 1 byte: length of instance name
	Instance    [63]byte         // 63 bytes: instance name
	FgPrismLen  uint8            // 1 byte: length of foreground prism name
	FgPrism     [63]byte         // 63 bytes: foreground prism name
	PrismCount  uint8            // 1 byte: number of active prisms
	_padding    [3]byte          // 3 bytes: padding for alignment
	Prisms      [16]PrismEntry   // 16 * 80 = 1280 bytes
}

func (s *PrismRuntimeState) GetInstance() string {
	return string(s.Instance[:s.InstanceLen])
}

func (s *PrismRuntimeState) SetInstance(name string) {
	if len(name) > NameMaxLen {
		name = name[:NameMaxLen]
	}
	s.InstanceLen = uint8(len(name))
	copy(s.Instance[:], name)
	for i := int(s.InstanceLen); i < NameMaxLen; i++ {
		s.Instance[i] = 0
	}
}

func (s *PrismRuntimeState) GetFgPrism() string {
	return string(s.FgPrism[:s.FgPrismLen])
}

func (s *PrismRuntimeState) SetFgPrism(name string) {
	if len(name) > NameMaxLen {
		name = name[:NameMaxLen]
	}
	s.FgPrismLen = uint8(len(name))
	copy(s.FgPrism[:], name)
	for i := int(s.FgPrismLen); i < NameMaxLen; i++ {
		s.FgPrism[i] = 0
	}
}

func (s *PrismRuntimeState) ActivePrisms() []PrismEntry {
	result := make([]PrismEntry, 0, s.PrismCount)
	for i := 0; i < int(s.PrismCount); i++ {
		if s.Prisms[i].IsActive() {
			result = append(result, s.Prisms[i])
		}
	}
	return result
}

type PanelEntry struct {
	InstanceLen uint8     // 1 byte: length of instance name
	Instance    [63]byte  // 63 bytes: instance name
	NameLen     uint8     // 1 byte: length of panel name
	Name        [63]byte  // 63 bytes: panel name
	PID         int32     // 4 bytes: prismctl process ID
	Healthy     uint8     // 1 byte: health status (0=unhealthy, 1=healthy)
	_padding    [3]byte   // 3 bytes: padding for alignment
}

func (e *PanelEntry) GetInstance() string {
	return string(e.Instance[:e.InstanceLen])
}

func (e *PanelEntry) SetInstance(name string) {
	if len(name) > NameMaxLen {
		name = name[:NameMaxLen]
	}
	e.InstanceLen = uint8(len(name))
	copy(e.Instance[:], name)
	for i := int(e.InstanceLen); i < NameMaxLen; i++ {
		e.Instance[i] = 0
	}
}

func (e *PanelEntry) GetName() string {
	return string(e.Name[:e.NameLen])
}

func (e *PanelEntry) SetName(name string) {
	if len(name) > NameMaxLen {
		name = name[:NameMaxLen]
	}
	e.NameLen = uint8(len(name))
	copy(e.Name[:], name)
	for i := int(e.NameLen); i < NameMaxLen; i++ {
		e.Name[i] = 0
	}
}

func (e *PanelEntry) IsHealthy() bool {
	return e.Healthy == 1
}

func (e *PanelEntry) IsActive() bool {
	return e.PID != 0
}

type ShinedState struct {
	Version    uint64           // 8 bytes: sequence counter (odd=writing, even=complete)
	PanelCount uint8            // 1 byte: number of active panels
	_padding   [7]byte          // 7 bytes: padding for alignment
	Panels     [32]PanelEntry   // 32 * 136 = 4352 bytes
}

func (s *ShinedState) ActivePanels() []PanelEntry {
	result := make([]PanelEntry, 0, s.PanelCount)
	for i := 0; i < int(s.PanelCount); i++ {
		if s.Panels[i].IsActive() {
			result = append(result, s.Panels[i])
		}
	}
	return result
}

func init() {
	if size := unsafe.Sizeof(PrismEntry{}); size != PrismEntrySize {
		panic(fmt.Sprintf("PrismEntry size mismatch: got %d, want %d", size, PrismEntrySize))
	}
	if size := unsafe.Sizeof(PrismRuntimeState{}); size != PrismRuntimeStateSize {
		panic(fmt.Sprintf("PrismRuntimeState size mismatch: got %d, want %d", size, PrismRuntimeStateSize))
	}
	if size := unsafe.Sizeof(PanelEntry{}); size != PanelEntrySize {
		panic(fmt.Sprintf("PanelEntry size mismatch: got %d, want %d", size, PanelEntrySize))
	}
	if size := unsafe.Sizeof(ShinedState{}); size != ShinedStateSize {
		panic(fmt.Sprintf("ShinedState size mismatch: got %d, want %d", size, ShinedStateSize))
	}
}
