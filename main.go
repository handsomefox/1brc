package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"slices"
	"sync"
	"syscall"
	"unsafe"
)

type Measurement struct {
	Total    int
	Sum      int64
	Min, Max int16
}

func (m *Measurement) Merge(value int16) {
	m.Total++
	m.Sum += int64(value)
	if value < m.Min {
		m.Min = value
	}
	if value > m.Max {
		m.Max = value
	}
}

func NewMeasurement(value int16) *Measurement {
	return &Measurement{
		Total: 1,
		Sum:   int64(value),
		Min:   value,
		Max:   value,
	}
}

func main() {
	debug.SetGCPercent(-1)
	measurements := run()

	keys := make([]string, 0, len(measurements))
	for k := range measurements {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	fmt.Print("{")
	for i, key := range keys {
		measure := measurements[key]
		fmt.Printf("%s=%.1f/%.1f/%.1f",
			key,
			formatTenths(measure.Min),
			formatAvg(measure.Sum, measure.Total),
			formatTenths(measure.Max),
		)
		if i != len(keys)-1 {
			fmt.Print(" ")
		}
	}
	fmt.Print("}\n")
}

func run() map[string]*Measurement {
	file, err := os.Open("measurements.txt")
	if err != nil {
		panic(err)
	}

	fileBytes := mmap(file)
	fileLength := len(fileBytes)

	workers := runtime.NumCPU()
	chunks := splitChunks(fileBytes, workers, fileLength)

	results := make([]map[string]*Measurement, len(chunks))
	var wg sync.WaitGroup
	wg.Add(len(chunks))

	for i, c := range chunks {
		i, c := i, c
		go func() {
			defer wg.Done()
			results[i] = parseChunk(fileBytes, c.start, c.end)
		}()
	}
	wg.Wait()

	return mergeWorkerMaps(results)
}

func mmap(f *os.File) []byte {
	fi, err := f.Stat()
	if err != nil {
		panic(err)
	}

	data, err := syscall.Mmap(
		int(f.Fd()),
		0,
		int(fi.Size()),
		syscall.PROT_READ,
		syscall.MAP_PRIVATE,
	)
	if err != nil {
		panic(err)
	}

	if err := syscall.Madvise(data, syscall.MADV_RANDOM); err != nil {
		panic(err)
	}

	return data
}

func parseFloat32(b []byte) int16 {
	i := 0
	sign := int16(1)
	if b[i] == '-' {
		sign = -1
		i++
	}

	var v int16
	for ; i < len(b) && b[i] != '.'; i++ {
		v = v*10 + int16(b[i]-'0')
	}
	i++

	if i < len(b) {
		v = v*10 + int16(b[i]-'0')
	}

	return sign * v
}

func unsafeBytesToString(b []byte) string {
	return unsafe.String(&b[0], len(b))
}

func formatTenths(v int16) float32 {
	return float32(v) / 10
}

func formatAvg(sum int64, total int) float32 {
	return float32(sum) / float32(total*10)
}

type chunk struct {
	start, end int
}

func splitChunks(fileBytes []byte, workers, fileLength int) []chunk {
	if workers < 1 {
		workers = 1
	}
	chunkSize := fileLength / workers

	chunks := make([]chunk, 0, workers)
	start := 0
	for i := 0; i < workers && start < fileLength; i++ {
		end := start + chunkSize
		if i == workers-1 || end > fileLength {
			end = fileLength
		} else {
			for end < fileLength && fileBytes[end-1] != '\n' {
				end++
			}
		}
		chunks = append(chunks, chunk{start: start, end: end})
		start = end
	}
	return chunks
}

func parseChunk(fileBytes []byte, start, end int) map[string]*Measurement {
	measurements := make(map[string]*Measurement, 10_000)
	i := start

	if i > 0 {
		for i < end && fileBytes[i-1] != '\n' {
			i++
		}
	}

	for i < end {
		cityStart := i
		for i < end && fileBytes[i] != ';' {
			i++
		}
		if i >= end {
			break
		}
		cityBytes := fileBytes[cityStart:i]
		i++ // skip ';'

		valStart := i
		for i < end && fileBytes[i] != '\n' {
			i++
		}
		valEnd := i
		if valEnd == valStart {
			if i < end {
				i++
			}
			continue
		}

		city := unsafeBytesToString(cityBytes)
		value := parseFloat32(fileBytes[valStart:valEnd])

		existing := measurements[city]
		if existing != nil {
			existing.Merge(value)
		} else {
			measurements[city] = NewMeasurement(value)
		}

		if i < end && fileBytes[i] == '\n' {
			i++
		}
	}

	return measurements
}

func mergeWorkerMaps(results []map[string]*Measurement) map[string]*Measurement {
	merged := make(map[string]*Measurement, 10_000)

	for _, data := range results {
		for city, m := range data {
			if existing, ok := merged[city]; ok {
				existing.Total += m.Total
				existing.Sum += m.Sum
				if m.Min < existing.Min {
					existing.Min = m.Min
				}
				if m.Max > existing.Max {
					existing.Max = m.Max
				}
			} else {
				merged[city] = m
			}
		}
	}

	return merged
}
