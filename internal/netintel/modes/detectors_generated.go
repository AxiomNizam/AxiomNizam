package modes

import "math"

func Detector001(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector002(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector003(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector004(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector005(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector006(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector007(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector008(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector009(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector010(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector011(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector012(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector013(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector014(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector015(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector016(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector017(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector018(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector019(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector020(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector021(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector022(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector023(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector024(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector025(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector026(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector027(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector028(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector029(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector030(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector031(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector032(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector033(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector034(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector035(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector036(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector037(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector038(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector039(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector040(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector041(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector042(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector043(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector044(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector045(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector046(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector047(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector048(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector049(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector050(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector051(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector052(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector053(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector054(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector055(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector056(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector057(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector058(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector059(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector060(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector061(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector062(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector063(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector064(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector065(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector066(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector067(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector068(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector069(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector070(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector071(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector072(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector073(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector074(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector075(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector076(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector077(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector078(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector079(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector080(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector081(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector082(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector083(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector084(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector085(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector086(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector087(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector088(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector089(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector090(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector091(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector092(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector093(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector094(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector095(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector096(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector097(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector098(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector099(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector100(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector101(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector102(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector103(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector104(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector105(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector106(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector107(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector108(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector109(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector110(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector111(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector112(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector113(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector114(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector115(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector116(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector117(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector118(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector119(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector120(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector121(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector122(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector123(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector124(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector125(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector126(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector127(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector128(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector129(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector130(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector131(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector132(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector133(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector134(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector135(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector136(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector137(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector138(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector139(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector140(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector141(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector142(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector143(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector144(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector145(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector146(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector147(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector148(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector149(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector150(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector151(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector152(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector153(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector154(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector155(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector156(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector157(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector158(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector159(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector160(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector161(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector162(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector163(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector164(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector165(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector166(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector167(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector168(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector169(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector170(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector171(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector172(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector173(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector174(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector175(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector176(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector177(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector178(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector179(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector180(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector181(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector182(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector183(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector184(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector185(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector186(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector187(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector188(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector189(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector190(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector191(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector192(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector193(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector194(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector195(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector196(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector197(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector198(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector199(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector200(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector201(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector202(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector203(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector204(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector205(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector206(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector207(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector208(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector209(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector210(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector211(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}

func Detector212(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.600
	return mean + sigma*weight
}

func Detector213(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.700
	return mean + sigma*weight
}

func Detector214(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.800
	return mean + sigma*weight
}

func Detector215(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.900
	return mean + sigma*weight
}

func Detector216(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.100
	return mean + sigma*weight
}

func Detector217(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.200
	return mean + sigma*weight
}

func Detector218(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.300
	return mean + sigma*weight
}

func Detector219(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.400
	return mean + sigma*weight
}

func Detector220(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := s - mean
		variance += d * d
	}
	variance = variance / float64(len(samples))
	sigma := math.Sqrt(variance)
	weight := 0.500
	return mean + sigma*weight
}
