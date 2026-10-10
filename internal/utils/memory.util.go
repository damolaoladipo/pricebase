package utils

import (
	"fmt"
	"runtime"
	"strings"
)

// MemoryStats represents Go runtime memory statistics.
type MemoryStats struct {
	Alloc         uint64  `json:"alloc"`
	TotalAlloc    uint64  `json:"totalAlloc"`
	Sys           uint64  `json:"sys"`
	Lookups       uint64  `json:"lookups"`
	Mallocs       uint64  `json:"mallocs"`
	Frees         uint64  `json:"frees"`
	HeapAlloc     uint64  `json:"heapAlloc"`
	HeapSys       uint64  `json:"heapSys"`
	HeapIdle      uint64  `json:"heapIdle"`
	HeapInuse     uint64  `json:"heapInuse"`
	HeapReleased  uint64  `json:"heapReleased"`
	HeapObjects   uint64  `json:"heapObjects"`
	StackInuse    uint64  `json:"stackInuse"`
	StackSys      uint64  `json:"stackSys"`
	MSpanInuse    uint64  `json:"mSpanInuse"`
	MSpanSys      uint64  `json:"mSpanSys"`
	MCacheInuse   uint64  `json:"mCacheInuse"`
	MCacheSys     uint64  `json:"mCacheSys"`
	BuckHashSys   uint64  `json:"buckHashSys"`
	GCSys         uint64  `json:"gcSys"`
	OtherSys      uint64  `json:"otherSys"`
	NextGC        uint64  `json:"nextGC"`
	LastGC        uint64  `json:"lastGC"`
	PauseTotalNs  uint64  `json:"pauseTotalNs"`
	NumGC         uint32  `json:"numGC"`
	NumForcedGC   uint32  `json:"numForcedGC"`
	GCCPUFraction float64 `json:"gcCPUFraction"`
}

// GetMemoryStats returns the current Go runtime memory statistics.
func GetMemoryStats() MemoryStats {
	var mem runtime.MemStats

	runtime.ReadMemStats(&mem)

	return MemoryStats{
		Alloc:         mem.Alloc,
		TotalAlloc:    mem.TotalAlloc,
		Sys:           mem.Sys,
		Lookups:       mem.Lookups,
		Mallocs:       mem.Mallocs,
		Frees:         mem.Frees,
		HeapAlloc:     mem.HeapAlloc,
		HeapSys:       mem.HeapSys,
		HeapIdle:      mem.HeapIdle,
		HeapInuse:     mem.HeapInuse,
		HeapReleased:  mem.HeapReleased,
		HeapObjects:   mem.HeapObjects,
		StackInuse:    mem.StackInuse,
		StackSys:      mem.StackSys,
		MSpanInuse:    mem.MSpanInuse,
		MSpanSys:      mem.MSpanSys,
		MCacheInuse:   mem.MCacheInuse,
		MCacheSys:     mem.MCacheSys,
		BuckHashSys:   mem.BuckHashSys,
		GCSys:         mem.GCSys,
		OtherSys:      mem.OtherSys,
		NextGC:        mem.NextGC,
		LastGC:        mem.LastGC,
		PauseTotalNs:  mem.PauseTotalNs,
		NumGC:         mem.NumGC,
		NumForcedGC:   mem.NumForcedGC,
		GCCPUFraction: mem.GCCPUFraction,
	}
}

// GetHeapSize returns the current Go heap size in the requested unit.
//
// Supported units:
// bytes
// kb, KB, kilobytes
// mb, MB, megabytes
// gb, GB, gigabytes
func GetHeapSize(unit string) any {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	total := float64(mem.HeapSys)

	switch strings.ToLower(unit) {
	case "bytes":
		return uint64(total)

	case "kb", "kilobytes":
		return fmt.Sprintf("%.2f", total/1024)

	case "mb", "megabytes":
		return fmt.Sprintf("%.2f", total/1024/1024)

	case "gb", "gigabytes":
		return fmt.Sprintf("%.2f", total/1024/1024/1024)

	default:
		return nil
	}
}
