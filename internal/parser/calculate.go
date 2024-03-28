package parser

import (
	"slices"

	"github.com/montanaflynn/stats"
)

const eps = 1e-9

type ExecutionDurationStats struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
	Avg float64 `json:"avg"`
	Med float64 `json:"med"`
	Std float64 `json:"std"`
}

func calcStats(d []float64) ExecutionDurationStats {
	if len(d) == 0 {
		return ExecutionDurationStats{}
	}
	min := slices.Min(d)
	max := slices.Max(d)
	avg, _ := stats.Mean(d)
	med, _ := stats.Median(d)
	std, _ := stats.StandardDeviation(d)

	return ExecutionDurationStats{
		Min: min,
		Max: max,
		Avg: avg,
		Med: med,
		Std: std,
	}

}

func adjustRate(rate float64) float64 {
	if rate < eps {
		return 0
	}
	return rate
}
