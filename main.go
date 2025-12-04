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
	Sum, Min, Max float64
	Total         int
}

func NewMeasurement(value float64) Measurement {
	return Measurement{Sum: value, Min: min(math.MaxFloat64, value), Max: max(-math.MaxFloat64, value), Total: 1}
}

func ParseLine(line string) (city string, value float64) {
	split := strings.SplitN(line, ";", 2)
	measureFloat, err := strconv.ParseFloat(split[1], 64)
	if err != nil {
		panic(err)
	}
	return split[0], measureFloat
}

func MergeMeasurement(existing Measurement, value float64) Measurement {
	existing.Total++
	existing.Sum += value
	existing.Min = min(existing.Min, value)
	existing.Max = max(existing.Max, value)
	return existing
}

func main() {
	f, err := os.Open("measurements.txt")
	if err != nil {
		panic(err)
	}
	sc := bufio.NewScanner(f)

	data := make(map[string]Measurement)

	for sc.Scan() {
		line := sc.Text()
		city, measurement := ParseLine(line)
		if existing, ok := data[city]; ok {
			data[city] = MergeMeasurement(existing, measurement)
		} else {
			data[city] = NewMeasurement(measurement)
		}
	}

	keys := slices.Collect(maps.Keys(data))
	slices.Sort(keys)

	print("{")
	for i, key := range keys {
		measure := data[key]
		fmt.Printf("%s=%.1f/%.1f/%.1f", key, measure.Min, measure.Sum/float64(measure.Total), measure.Max)
		if i != len(keys)-1 {
			print(" ")
		}
	}
	print("}\n")
}
