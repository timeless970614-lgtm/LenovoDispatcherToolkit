//go:build windows

package power

// ── Windows Driver Structures ────────────────────────────────────────────

// MSRIO matches the kernel driver's MSR_IO structure
type MSRIO struct {
	Index  uint32
	Value  uint64
	Status int32
}

// PCICfgIO matches the kernel driver's PCI_CFG_IO structure
type PCICfgIO struct {
	Bus    uint32
	Dev    uint32
	Func   uint32
	Offset uint32
	Size   uint32
	Value  uint32
	Status int32
}

// PhysMemIO matches the kernel driver's PHYS_MEM_IO structure
type PhysMemIO struct {
	PhysAddr uint64
	Size     uint32
	Data     [256]byte
	Status   int32
}