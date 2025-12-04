package main

import (
	"bytes"
	"fmt"
	"maps"
	"math"
	"os"
	"slices"
	"strconv"
	"syscall"
)

type Measurement struct {
	Total         int
	Sum, Min, Max float32
}

func (m *Measurement) Merge(value float32) {
	m.Total++
	m.Sum += value
	m.Min = min(m.Min, value)
	m.Max = max(m.Max, value)
}

func NewMeasurement(value float32) *Measurement {
	return &Measurement{
		Sum:   value,
		Min:   min(math.MaxFloat32, value),
		Max:   max(-math.MaxFloat32, value),
		Total: 1,
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
		nlBytes        = []byte{'\n'}
	)

	for fileReadOffset < len(fileBytes) {
		nl := bytes.IndexByte(fileBytes[fileReadOffset:], '\n')

		var line []byte
		if nl == -1 {
			line = fileBytes[fileReadOffset:]
			fileReadOffset = len(fileBytes)
		} else {
			line = fileBytes[fileReadOffset : fileReadOffset+nl]
			fileReadOffset += nl + 1
		}
		if len(line) == 0 || bytes.Equal(line, nlBytes) {
			continue
		}

		sep := bytes.IndexByte(line, ';')
		cityBytes := line[:sep]
		valBytes := line[sep+1:]
		val, err := strconv.ParseFloat(string(valBytes), 32)
		if err != nil {
			panic(err)
		}

		city, measurement := string(cityBytes), float32(val)
		if existing, ok := measurements[city]; ok {
			existing.Merge(measurement)
		} else {
			measurements[city] = NewMeasurement(measurement)
		}
	}

	keys := slices.Collect(maps.Keys(measurements))
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
