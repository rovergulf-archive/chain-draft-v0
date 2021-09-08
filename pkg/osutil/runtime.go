package osutil

import "runtime"

func GetAvailableMemory() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Frees
}

func GetAvailableCPUs() int {
	return runtime.NumCPU()
}
