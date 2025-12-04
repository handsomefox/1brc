package main

import (
	"fmt"
	"os"
	"slices"
	"syscall"
	"unsafe"
)

type Measurement struct {
	Total         int
	Sum, Min, Max float32
}

func (m *Measurement) Merge(value float32) {
	m.Total++
	m.Sum += value
	if value < m.Min {
		m.Min = value
	}
	if value > m.Max {
		m.Max = value
	}
}

func NewMeasurement(value float32) *Measurement {
	return &Measurement{
		Total: 1,
		Sum:   value,
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
		fmt.Printf("%s=%.1f/%.1f/%.1f", key, measure.Min, measure.Sum/float32(measure.Total), measure.Max)
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

func parseFloat32(b []byte) float32 {
	if len(b) == 0 {
		return 0
	}
	sign := float32(1)
	i := 0
	if b[0] == '-' {
		sign = -1
		i++
	}

	var intPart int32
	for i < len(b) && b[i] != '.' {
		intPart = intPart*10 + int32(b[i]-'0')
		i++
	}

	var fracPart int32
	var fracDiv float32 = 1

	if i < len(b) && b[i] == '.' {
		i++
		for i < len(b) {
			fracPart = fracPart*10 + int32(b[i]-'0')
			fracDiv *= 10
			i++
		}
	}

	return sign * (float32(intPart) + float32(fracPart)/fracDiv)
}

func unsafeBytesToString(b []byte) string {
	return unsafe.String(&b[0], len(b))
}
