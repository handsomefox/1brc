package main

import (
	"fmt"
	"os"
	"slices"
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
	file, err := os.Open("measurements.txt")
	if err != nil {
		panic(err)
	}

	var (
		fileBytes      = mmap(file)
		fileReadOffset = 0
		measurements   = make(map[string]*Measurement, 10_000)
	)

	for fileReadOffset < len(fileBytes) {
		i := fileReadOffset

		start := i
		for i < len(fileBytes) && fileBytes[i] != ';' {
			i++
		}
		if i >= len(fileBytes) {
			break
		}
		cityBytes := fileBytes[start:i]
		i++ // skip ';'

		valStart := i
		for i < len(fileBytes) && fileBytes[i] != '\n' {
			i++
		}
		valEnd := i
		if valEnd == valStart { // empty value
			if i < len(fileBytes) {
				i++
			}
			fileReadOffset = i
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

		// skip '\n' if present
		if i < len(fileBytes) && fileBytes[i] == '\n' {
			i++
		}
		fileReadOffset = i
	}

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

	if err := syscall.Madvise(data, syscall.MADV_SEQUENTIAL); err != nil {
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
