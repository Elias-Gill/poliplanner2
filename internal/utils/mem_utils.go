package utils

import (
	"runtime"

	"github.com/elias-gill/poliplanner2/internal/logger"
)

func MemUsageStatus(label string, do func()) {
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// ---- Tested function ----
	do()
	// -------------------------

	runtime.ReadMemStats(&m2)

	allocKB := float64(m2.TotalAlloc-m1.TotalAlloc) / 1024
	allocMB := allocKB / 1024

	logger.Debug(label+" - Alloc delta",
		"KB", allocKB,
		"MB", allocMB,
	)

	heapKB := float64(m2.HeapAlloc-m1.HeapAlloc) / 1024
	heapMB := heapKB / 1024

	logger.Debug(label+" - Heap delta",
		"KB", heapKB,
		"MB", heapMB,
	)
}
