package main

import (
	"bufio"
	"fmt"
	"maps"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
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
	return &Measurement{Sum: value, Min: min(math.MaxFloat32, value), Max: max(-math.MaxFloat32, value), Total: 1}
}

func ParseLine(line string) (city string, value float32) {
	split := strings.SplitN(line, ";", 2)
	measureFloat, err := strconv.ParseFloat(split[1], 32)
	if err != nil {
		panic(err)
	}
	return split[0], float32(measureFloat)
}

func main() {
	f, err := os.Open("measurements.txt")
	if err != nil {
		panic(err)
	}
	sc := bufio.NewScanner(f)

	data := make(map[string]*Measurement)

	for sc.Scan() {
		line := sc.Text()
		city, measurement := ParseLine(line)
		if existing, ok := data[city]; ok {
			existing.Merge(measurement)
		} else {
			data[city] = NewMeasurement(measurement)
		}
	}

	keys := slices.Collect(maps.Keys(data))
	slices.Sort(keys)

	print("{")
	for i, key := range keys {
		measure := data[key]
		fmt.Printf("%s=%.1f/%.1f/%.1f", key, measure.Min, measure.Sum/float32(measure.Total), measure.Max)
		if i != len(keys)-1 {
			print(" ")
		}
	}
	print("}\n")
}
