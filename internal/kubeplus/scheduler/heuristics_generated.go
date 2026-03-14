package scheduler

func HeuristicScore001(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 2
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((2 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (2 + 3)
	}
	if n.Zone != "" {
		score += (2 % 5)
	}
	return score
}

func HeuristicScore002(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 3
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((3 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (3 + 3)
	}
	if n.Zone != "" {
		score += (3 % 5)
	}
	return score
}

func HeuristicScore003(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 4
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((4 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (4 + 3)
	}
	if n.Zone != "" {
		score += (4 % 5)
	}
	return score
}

func HeuristicScore004(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 5
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((5 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (5 + 3)
	}
	if n.Zone != "" {
		score += (5 % 5)
	}
	return score
}

func HeuristicScore005(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 6
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((6 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (6 + 3)
	}
	if n.Zone != "" {
		score += (6 % 5)
	}
	return score
}

func HeuristicScore006(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 7
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((7 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (7 + 3)
	}
	if n.Zone != "" {
		score += (7 % 5)
	}
	return score
}

func HeuristicScore007(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 8
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((8 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (8 + 3)
	}
	if n.Zone != "" {
		score += (8 % 5)
	}
	return score
}

func HeuristicScore008(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 9
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((9 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (9 + 3)
	}
	if n.Zone != "" {
		score += (9 % 5)
	}
	return score
}

func HeuristicScore009(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 10
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((10 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (10 + 3)
	}
	if n.Zone != "" {
		score += (10 % 5)
	}
	return score
}

func HeuristicScore010(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 11
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((11 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (11 + 3)
	}
	if n.Zone != "" {
		score += (11 % 5)
	}
	return score
}

func HeuristicScore011(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 12
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((12 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (12 + 3)
	}
	if n.Zone != "" {
		score += (12 % 5)
	}
	return score
}

func HeuristicScore012(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 13
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((13 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (13 + 3)
	}
	if n.Zone != "" {
		score += (13 % 5)
	}
	return score
}

func HeuristicScore013(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 14
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((14 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (14 + 3)
	}
	if n.Zone != "" {
		score += (14 % 5)
	}
	return score
}

func HeuristicScore014(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 15
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((15 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (15 + 3)
	}
	if n.Zone != "" {
		score += (15 % 5)
	}
	return score
}

func HeuristicScore015(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 16
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((16 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (16 + 3)
	}
	if n.Zone != "" {
		score += (16 % 5)
	}
	return score
}

func HeuristicScore016(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 17
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((17 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (17 + 3)
	}
	if n.Zone != "" {
		score += (17 % 5)
	}
	return score
}

func HeuristicScore017(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 1
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((1 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (1 + 3)
	}
	if n.Zone != "" {
		score += (1 % 5)
	}
	return score
}

func HeuristicScore018(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 2
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((2 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (2 + 3)
	}
	if n.Zone != "" {
		score += (2 % 5)
	}
	return score
}

func HeuristicScore019(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 3
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((3 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (3 + 3)
	}
	if n.Zone != "" {
		score += (3 % 5)
	}
	return score
}

func HeuristicScore020(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 4
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((4 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (4 + 3)
	}
	if n.Zone != "" {
		score += (4 % 5)
	}
	return score
}

func HeuristicScore021(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 5
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((5 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (5 + 3)
	}
	if n.Zone != "" {
		score += (5 % 5)
	}
	return score
}

func HeuristicScore022(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 6
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((6 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (6 + 3)
	}
	if n.Zone != "" {
		score += (6 % 5)
	}
	return score
}

func HeuristicScore023(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 7
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((7 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (7 + 3)
	}
	if n.Zone != "" {
		score += (7 % 5)
	}
	return score
}

func HeuristicScore024(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 8
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((8 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (8 + 3)
	}
	if n.Zone != "" {
		score += (8 % 5)
	}
	return score
}

func HeuristicScore025(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 9
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((9 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (9 + 3)
	}
	if n.Zone != "" {
		score += (9 % 5)
	}
	return score
}

func HeuristicScore026(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 10
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((10 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (10 + 3)
	}
	if n.Zone != "" {
		score += (10 % 5)
	}
	return score
}

func HeuristicScore027(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 11
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((11 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (11 + 3)
	}
	if n.Zone != "" {
		score += (11 % 5)
	}
	return score
}

func HeuristicScore028(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 12
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((12 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (12 + 3)
	}
	if n.Zone != "" {
		score += (12 % 5)
	}
	return score
}

func HeuristicScore029(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 13
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((13 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (13 + 3)
	}
	if n.Zone != "" {
		score += (13 % 5)
	}
	return score
}

func HeuristicScore030(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 14
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((14 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (14 + 3)
	}
	if n.Zone != "" {
		score += (14 % 5)
	}
	return score
}

func HeuristicScore031(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 15
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((15 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (15 + 3)
	}
	if n.Zone != "" {
		score += (15 % 5)
	}
	return score
}

func HeuristicScore032(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 16
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((16 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (16 + 3)
	}
	if n.Zone != "" {
		score += (16 % 5)
	}
	return score
}

func HeuristicScore033(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 17
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((17 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (17 + 3)
	}
	if n.Zone != "" {
		score += (17 % 5)
	}
	return score
}

func HeuristicScore034(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 1
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((1 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (1 + 3)
	}
	if n.Zone != "" {
		score += (1 % 5)
	}
	return score
}

func HeuristicScore035(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 2
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((2 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (2 + 3)
	}
	if n.Zone != "" {
		score += (2 % 5)
	}
	return score
}

func HeuristicScore036(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 3
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((3 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (3 + 3)
	}
	if n.Zone != "" {
		score += (3 % 5)
	}
	return score
}

func HeuristicScore037(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 4
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((4 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (4 + 3)
	}
	if n.Zone != "" {
		score += (4 % 5)
	}
	return score
}

func HeuristicScore038(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 5
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((5 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (5 + 3)
	}
	if n.Zone != "" {
		score += (5 % 5)
	}
	return score
}

func HeuristicScore039(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 6
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((6 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (6 + 3)
	}
	if n.Zone != "" {
		score += (6 % 5)
	}
	return score
}

func HeuristicScore040(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 7
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((7 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (7 + 3)
	}
	if n.Zone != "" {
		score += (7 % 5)
	}
	return score
}

func HeuristicScore041(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 8
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((8 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (8 + 3)
	}
	if n.Zone != "" {
		score += (8 % 5)
	}
	return score
}

func HeuristicScore042(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 9
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((9 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (9 + 3)
	}
	if n.Zone != "" {
		score += (9 % 5)
	}
	return score
}

func HeuristicScore043(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 10
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((10 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (10 + 3)
	}
	if n.Zone != "" {
		score += (10 % 5)
	}
	return score
}

func HeuristicScore044(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 11
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((11 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (11 + 3)
	}
	if n.Zone != "" {
		score += (11 % 5)
	}
	return score
}

func HeuristicScore045(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 12
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((12 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (12 + 3)
	}
	if n.Zone != "" {
		score += (12 % 5)
	}
	return score
}

func HeuristicScore046(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 13
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((13 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (13 + 3)
	}
	if n.Zone != "" {
		score += (13 % 5)
	}
	return score
}

func HeuristicScore047(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 14
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((14 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (14 + 3)
	}
	if n.Zone != "" {
		score += (14 % 5)
	}
	return score
}

func HeuristicScore048(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 15
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((15 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (15 + 3)
	}
	if n.Zone != "" {
		score += (15 % 5)
	}
	return score
}

func HeuristicScore049(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 16
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((16 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (16 + 3)
	}
	if n.Zone != "" {
		score += (16 % 5)
	}
	return score
}

func HeuristicScore050(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 17
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((17 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (17 + 3)
	}
	if n.Zone != "" {
		score += (17 % 5)
	}
	return score
}

func HeuristicScore051(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 1
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((1 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (1 + 3)
	}
	if n.Zone != "" {
		score += (1 % 5)
	}
	return score
}

func HeuristicScore052(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 2
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((2 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (2 + 3)
	}
	if n.Zone != "" {
		score += (2 % 5)
	}
	return score
}

func HeuristicScore053(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 3
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((3 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (3 + 3)
	}
	if n.Zone != "" {
		score += (3 % 5)
	}
	return score
}

func HeuristicScore054(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 4
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((4 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (4 + 3)
	}
	if n.Zone != "" {
		score += (4 % 5)
	}
	return score
}

func HeuristicScore055(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 5
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((5 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (5 + 3)
	}
	if n.Zone != "" {
		score += (5 % 5)
	}
	return score
}

func HeuristicScore056(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 6
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((6 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (6 + 3)
	}
	if n.Zone != "" {
		score += (6 % 5)
	}
	return score
}

func HeuristicScore057(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 7
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((7 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (7 + 3)
	}
	if n.Zone != "" {
		score += (7 % 5)
	}
	return score
}

func HeuristicScore058(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 8
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((8 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (8 + 3)
	}
	if n.Zone != "" {
		score += (8 % 5)
	}
	return score
}

func HeuristicScore059(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 9
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((9 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (9 + 3)
	}
	if n.Zone != "" {
		score += (9 % 5)
	}
	return score
}

func HeuristicScore060(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 10
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((10 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (10 + 3)
	}
	if n.Zone != "" {
		score += (10 % 5)
	}
	return score
}

func HeuristicScore061(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 11
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((11 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (11 + 3)
	}
	if n.Zone != "" {
		score += (11 % 5)
	}
	return score
}

func HeuristicScore062(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 12
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((12 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (12 + 3)
	}
	if n.Zone != "" {
		score += (12 % 5)
	}
	return score
}

func HeuristicScore063(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 13
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((13 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (13 + 3)
	}
	if n.Zone != "" {
		score += (13 % 5)
	}
	return score
}

func HeuristicScore064(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 14
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((14 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (14 + 3)
	}
	if n.Zone != "" {
		score += (14 % 5)
	}
	return score
}

func HeuristicScore065(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 15
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((15 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (15 + 3)
	}
	if n.Zone != "" {
		score += (15 % 5)
	}
	return score
}

func HeuristicScore066(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 16
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((16 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (16 + 3)
	}
	if n.Zone != "" {
		score += (16 % 5)
	}
	return score
}

func HeuristicScore067(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 17
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((17 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (17 + 3)
	}
	if n.Zone != "" {
		score += (17 % 5)
	}
	return score
}

func HeuristicScore068(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 1
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((1 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (1 + 3)
	}
	if n.Zone != "" {
		score += (1 % 5)
	}
	return score
}

func HeuristicScore069(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 2
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((2 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (2 + 3)
	}
	if n.Zone != "" {
		score += (2 % 5)
	}
	return score
}

func HeuristicScore070(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 3
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((3 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (3 + 3)
	}
	if n.Zone != "" {
		score += (3 % 5)
	}
	return score
}

func HeuristicScore071(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 4
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((4 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (4 + 3)
	}
	if n.Zone != "" {
		score += (4 % 5)
	}
	return score
}

func HeuristicScore072(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 5
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((5 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (5 + 3)
	}
	if n.Zone != "" {
		score += (5 % 5)
	}
	return score
}

func HeuristicScore073(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 6
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((6 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (6 + 3)
	}
	if n.Zone != "" {
		score += (6 % 5)
	}
	return score
}

func HeuristicScore074(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 7
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((7 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (7 + 3)
	}
	if n.Zone != "" {
		score += (7 % 5)
	}
	return score
}

func HeuristicScore075(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 8
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((8 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (8 + 3)
	}
	if n.Zone != "" {
		score += (8 % 5)
	}
	return score
}

func HeuristicScore076(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 9
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((9 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (9 + 3)
	}
	if n.Zone != "" {
		score += (9 % 5)
	}
	return score
}

func HeuristicScore077(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 10
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((10 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (10 + 3)
	}
	if n.Zone != "" {
		score += (10 % 5)
	}
	return score
}

func HeuristicScore078(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 11
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((11 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (11 + 3)
	}
	if n.Zone != "" {
		score += (11 % 5)
	}
	return score
}

func HeuristicScore079(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 12
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((12 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (12 + 3)
	}
	if n.Zone != "" {
		score += (12 % 5)
	}
	return score
}

func HeuristicScore080(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 13
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((13 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (13 + 3)
	}
	if n.Zone != "" {
		score += (13 % 5)
	}
	return score
}

func HeuristicScore081(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 14
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((14 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (14 + 3)
	}
	if n.Zone != "" {
		score += (14 % 5)
	}
	return score
}

func HeuristicScore082(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 15
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((15 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (15 + 3)
	}
	if n.Zone != "" {
		score += (15 % 5)
	}
	return score
}

func HeuristicScore083(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 16
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((16 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (16 + 3)
	}
	if n.Zone != "" {
		score += (16 % 5)
	}
	return score
}

func HeuristicScore084(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 17
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((17 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (17 + 3)
	}
	if n.Zone != "" {
		score += (17 % 5)
	}
	return score
}

func HeuristicScore085(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 1
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((1 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (1 + 3)
	}
	if n.Zone != "" {
		score += (1 % 5)
	}
	return score
}

func HeuristicScore086(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 2
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((2 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (2 + 3)
	}
	if n.Zone != "" {
		score += (2 % 5)
	}
	return score
}

func HeuristicScore087(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 3
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((3 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (3 + 3)
	}
	if n.Zone != "" {
		score += (3 % 5)
	}
	return score
}

func HeuristicScore088(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 4
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((4 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (4 + 3)
	}
	if n.Zone != "" {
		score += (4 % 5)
	}
	return score
}

func HeuristicScore089(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 5
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((5 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (5 + 3)
	}
	if n.Zone != "" {
		score += (5 % 5)
	}
	return score
}

func HeuristicScore090(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 6
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((6 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (6 + 3)
	}
	if n.Zone != "" {
		score += (6 % 5)
	}
	return score
}

func HeuristicScore091(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 7
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((7 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (7 + 3)
	}
	if n.Zone != "" {
		score += (7 % 5)
	}
	return score
}

func HeuristicScore092(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 8
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((8 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (8 + 3)
	}
	if n.Zone != "" {
		score += (8 % 5)
	}
	return score
}

func HeuristicScore093(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 9
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((9 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (9 + 3)
	}
	if n.Zone != "" {
		score += (9 % 5)
	}
	return score
}

func HeuristicScore094(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 10
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((10 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (10 + 3)
	}
	if n.Zone != "" {
		score += (10 % 5)
	}
	return score
}

func HeuristicScore095(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 11
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((11 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (11 + 3)
	}
	if n.Zone != "" {
		score += (11 % 5)
	}
	return score
}

func HeuristicScore096(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 12
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((12 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (12 + 3)
	}
	if n.Zone != "" {
		score += (12 % 5)
	}
	return score
}

func HeuristicScore097(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 13
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((13 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (13 + 3)
	}
	if n.Zone != "" {
		score += (13 % 5)
	}
	return score
}

func HeuristicScore098(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 14
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((14 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (14 + 3)
	}
	if n.Zone != "" {
		score += (14 % 5)
	}
	return score
}

func HeuristicScore099(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 15
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((15 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (15 + 3)
	}
	if n.Zone != "" {
		score += (15 % 5)
	}
	return score
}

func HeuristicScore100(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 16
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((16 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (16 + 3)
	}
	if n.Zone != "" {
		score += (16 % 5)
	}
	return score
}

func HeuristicScore101(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 17
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((17 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (17 + 3)
	}
	if n.Zone != "" {
		score += (17 % 5)
	}
	return score
}

func HeuristicScore102(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 1
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((1 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (1 + 3)
	}
	if n.Zone != "" {
		score += (1 % 5)
	}
	return score
}

func HeuristicScore103(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 2
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((2 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (2 + 3)
	}
	if n.Zone != "" {
		score += (2 % 5)
	}
	return score
}

func HeuristicScore104(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 3
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((3 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (3 + 3)
	}
	if n.Zone != "" {
		score += (3 % 5)
	}
	return score
}

func HeuristicScore105(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 4
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((4 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (4 + 3)
	}
	if n.Zone != "" {
		score += (4 % 5)
	}
	return score
}

func HeuristicScore106(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 5
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((5 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (5 + 3)
	}
	if n.Zone != "" {
		score += (5 % 5)
	}
	return score
}

func HeuristicScore107(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 6
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((6 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (6 + 3)
	}
	if n.Zone != "" {
		score += (6 % 5)
	}
	return score
}

func HeuristicScore108(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 7
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((7 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (7 + 3)
	}
	if n.Zone != "" {
		score += (7 % 5)
	}
	return score
}

func HeuristicScore109(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 8
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((8 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (8 + 3)
	}
	if n.Zone != "" {
		score += (8 % 5)
	}
	return score
}

func HeuristicScore110(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 9
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((9 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (9 + 3)
	}
	if n.Zone != "" {
		score += (9 % 5)
	}
	return score
}

func HeuristicScore111(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 10
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((10 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (10 + 3)
	}
	if n.Zone != "" {
		score += (10 % 5)
	}
	return score
}

func HeuristicScore112(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 11
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((11 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (11 + 3)
	}
	if n.Zone != "" {
		score += (11 % 5)
	}
	return score
}

func HeuristicScore113(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 12
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((12 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (12 + 3)
	}
	if n.Zone != "" {
		score += (12 % 5)
	}
	return score
}

func HeuristicScore114(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 13
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((13 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (13 + 3)
	}
	if n.Zone != "" {
		score += (13 % 5)
	}
	return score
}

func HeuristicScore115(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 14
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((14 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (14 + 3)
	}
	if n.Zone != "" {
		score += (14 % 5)
	}
	return score
}

func HeuristicScore116(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 15
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((15 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (15 + 3)
	}
	if n.Zone != "" {
		score += (15 % 5)
	}
	return score
}

func HeuristicScore117(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 16
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((16 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (16 + 3)
	}
	if n.Zone != "" {
		score += (16 % 5)
	}
	return score
}

func HeuristicScore118(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 17
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((17 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (17 + 3)
	}
	if n.Zone != "" {
		score += (17 % 5)
	}
	return score
}

func HeuristicScore119(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 1
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((1 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (1 + 3)
	}
	if n.Zone != "" {
		score += (1 % 5)
	}
	return score
}

func HeuristicScore120(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 2
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((2 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (2 + 3)
	}
	if n.Zone != "" {
		score += (2 % 5)
	}
	return score
}

func HeuristicScore121(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 3
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((3 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (3 + 3)
	}
	if n.Zone != "" {
		score += (3 % 5)
	}
	return score
}

func HeuristicScore122(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 4
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((4 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (4 + 3)
	}
	if n.Zone != "" {
		score += (4 % 5)
	}
	return score
}

func HeuristicScore123(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 5
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((5 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (5 + 3)
	}
	if n.Zone != "" {
		score += (5 % 5)
	}
	return score
}

func HeuristicScore124(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 6
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((6 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (6 + 3)
	}
	if n.Zone != "" {
		score += (6 % 5)
	}
	return score
}

func HeuristicScore125(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 7
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((7 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (7 + 3)
	}
	if n.Zone != "" {
		score += (7 % 5)
	}
	return score
}

func HeuristicScore126(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 8
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((8 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (8 + 3)
	}
	if n.Zone != "" {
		score += (8 % 5)
	}
	return score
}

func HeuristicScore127(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 9
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((9 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (9 + 3)
	}
	if n.Zone != "" {
		score += (9 % 5)
	}
	return score
}

func HeuristicScore128(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 10
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((10 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (10 + 3)
	}
	if n.Zone != "" {
		score += (10 % 5)
	}
	return score
}

func HeuristicScore129(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 11
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((11 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (11 + 3)
	}
	if n.Zone != "" {
		score += (11 % 5)
	}
	return score
}

func HeuristicScore130(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 12
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((12 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (12 + 3)
	}
	if n.Zone != "" {
		score += (12 % 5)
	}
	return score
}

func HeuristicScore131(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 13
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((13 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (13 + 3)
	}
	if n.Zone != "" {
		score += (13 % 5)
	}
	return score
}

func HeuristicScore132(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 14
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((14 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (14 + 3)
	}
	if n.Zone != "" {
		score += (14 % 5)
	}
	return score
}

func HeuristicScore133(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 15
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((15 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (15 + 3)
	}
	if n.Zone != "" {
		score += (15 % 5)
	}
	return score
}

func HeuristicScore134(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 16
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((16 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (16 + 3)
	}
	if n.Zone != "" {
		score += (16 % 5)
	}
	return score
}

func HeuristicScore135(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 17
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((17 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (17 + 3)
	}
	if n.Zone != "" {
		score += (17 % 5)
	}
	return score
}

func HeuristicScore136(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 1
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((1 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (1 + 3)
	}
	if n.Zone != "" {
		score += (1 % 5)
	}
	return score
}

func HeuristicScore137(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 2
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((2 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (2 + 3)
	}
	if n.Zone != "" {
		score += (2 % 5)
	}
	return score
}

func HeuristicScore138(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 3
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((3 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (3 + 3)
	}
	if n.Zone != "" {
		score += (3 % 5)
	}
	return score
}

func HeuristicScore139(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 4
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((4 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (4 + 3)
	}
	if n.Zone != "" {
		score += (4 % 5)
	}
	return score
}

func HeuristicScore140(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 5
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((5 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (5 + 3)
	}
	if n.Zone != "" {
		score += (5 % 5)
	}
	return score
}

func HeuristicScore141(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 6
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((6 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (6 + 3)
	}
	if n.Zone != "" {
		score += (6 % 5)
	}
	return score
}

func HeuristicScore142(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 7
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((7 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (7 + 3)
	}
	if n.Zone != "" {
		score += (7 % 5)
	}
	return score
}

func HeuristicScore143(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 8
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((8 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (8 + 3)
	}
	if n.Zone != "" {
		score += (8 % 5)
	}
	return score
}

func HeuristicScore144(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 9
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((9 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (9 + 3)
	}
	if n.Zone != "" {
		score += (9 % 5)
	}
	return score
}

func HeuristicScore145(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 10
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((10 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (10 + 3)
	}
	if n.Zone != "" {
		score += (10 % 5)
	}
	return score
}

func HeuristicScore146(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 11
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((11 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (11 + 3)
	}
	if n.Zone != "" {
		score += (11 % 5)
	}
	return score
}

func HeuristicScore147(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 12
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((12 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (12 + 3)
	}
	if n.Zone != "" {
		score += (12 % 5)
	}
	return score
}

func HeuristicScore148(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 13
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((13 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (13 + 3)
	}
	if n.Zone != "" {
		score += (13 % 5)
	}
	return score
}

func HeuristicScore149(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 14
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((14 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (14 + 3)
	}
	if n.Zone != "" {
		score += (14 % 5)
	}
	return score
}

func HeuristicScore150(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 15
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((15 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (15 + 3)
	}
	if n.Zone != "" {
		score += (15 % 5)
	}
	return score
}

func HeuristicScore151(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 16
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((16 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (16 + 3)
	}
	if n.Zone != "" {
		score += (16 % 5)
	}
	return score
}

func HeuristicScore152(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 17
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((17 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (17 + 3)
	}
	if n.Zone != "" {
		score += (17 % 5)
	}
	return score
}

func HeuristicScore153(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 1
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((1 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (1 + 3)
	}
	if n.Zone != "" {
		score += (1 % 5)
	}
	return score
}

func HeuristicScore154(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 2
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((2 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (2 + 3)
	}
	if n.Zone != "" {
		score += (2 % 5)
	}
	return score
}

func HeuristicScore155(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 3
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((3 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (3 + 3)
	}
	if n.Zone != "" {
		score += (3 % 5)
	}
	return score
}

func HeuristicScore156(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 4
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((4 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (4 + 3)
	}
	if n.Zone != "" {
		score += (4 % 5)
	}
	return score
}

func HeuristicScore157(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 5
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((5 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (5 + 3)
	}
	if n.Zone != "" {
		score += (5 % 5)
	}
	return score
}

func HeuristicScore158(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 6
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((6 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (6 + 3)
	}
	if n.Zone != "" {
		score += (6 % 5)
	}
	return score
}

func HeuristicScore159(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 7
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((7 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (7 + 3)
	}
	if n.Zone != "" {
		score += (7 % 5)
	}
	return score
}

func HeuristicScore160(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 8
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((8 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (8 + 3)
	}
	if n.Zone != "" {
		score += (8 % 5)
	}
	return score
}

func HeuristicScore161(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 9
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((9 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (9 + 3)
	}
	if n.Zone != "" {
		score += (9 % 5)
	}
	return score
}

func HeuristicScore162(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 10
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((10 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (10 + 3)
	}
	if n.Zone != "" {
		score += (10 % 5)
	}
	return score
}

func HeuristicScore163(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 11
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((11 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (11 + 3)
	}
	if n.Zone != "" {
		score += (11 % 5)
	}
	return score
}

func HeuristicScore164(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 12
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((12 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (12 + 3)
	}
	if n.Zone != "" {
		score += (12 % 5)
	}
	return score
}

func HeuristicScore165(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 13
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((13 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (13 + 3)
	}
	if n.Zone != "" {
		score += (13 % 5)
	}
	return score
}

func HeuristicScore166(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 14
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((14 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (14 + 3)
	}
	if n.Zone != "" {
		score += (14 % 5)
	}
	return score
}

func HeuristicScore167(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 15
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((15 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (15 + 3)
	}
	if n.Zone != "" {
		score += (15 % 5)
	}
	return score
}

func HeuristicScore168(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 16
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((16 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (16 + 3)
	}
	if n.Zone != "" {
		score += (16 % 5)
	}
	return score
}

func HeuristicScore169(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 17
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((17 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (17 + 3)
	}
	if n.Zone != "" {
		score += (17 % 5)
	}
	return score
}

func HeuristicScore170(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 1
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((1 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (1 + 3)
	}
	if n.Zone != "" {
		score += (1 % 5)
	}
	return score
}

func HeuristicScore171(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 2
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((2 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (2 + 3)
	}
	if n.Zone != "" {
		score += (2 % 5)
	}
	return score
}

func HeuristicScore172(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 3
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((3 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (3 + 3)
	}
	if n.Zone != "" {
		score += (3 % 5)
	}
	return score
}

func HeuristicScore173(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 4
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((4 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (4 + 3)
	}
	if n.Zone != "" {
		score += (4 % 5)
	}
	return score
}

func HeuristicScore174(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 5
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((5 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (5 + 3)
	}
	if n.Zone != "" {
		score += (5 % 5)
	}
	return score
}

func HeuristicScore175(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 6
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((6 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (6 + 3)
	}
	if n.Zone != "" {
		score += (6 % 5)
	}
	return score
}

func HeuristicScore176(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 7
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((7 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (7 + 3)
	}
	if n.Zone != "" {
		score += (7 % 5)
	}
	return score
}

func HeuristicScore177(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 8
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((8 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (8 + 3)
	}
	if n.Zone != "" {
		score += (8 % 5)
	}
	return score
}

func HeuristicScore178(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 9
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((9 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (9 + 3)
	}
	if n.Zone != "" {
		score += (9 % 5)
	}
	return score
}

func HeuristicScore179(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 10
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((10 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (10 + 3)
	}
	if n.Zone != "" {
		score += (10 % 5)
	}
	return score
}

func HeuristicScore180(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 11
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((11 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (11 + 3)
	}
	if n.Zone != "" {
		score += (11 % 5)
	}
	return score
}

func HeuristicScore181(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 12
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((12 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (12 + 3)
	}
	if n.Zone != "" {
		score += (12 % 5)
	}
	return score
}

func HeuristicScore182(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 13
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((13 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (13 + 3)
	}
	if n.Zone != "" {
		score += (13 % 5)
	}
	return score
}

func HeuristicScore183(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 14
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((14 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (14 + 3)
	}
	if n.Zone != "" {
		score += (14 % 5)
	}
	return score
}

func HeuristicScore184(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 15
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((15 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (15 + 3)
	}
	if n.Zone != "" {
		score += (15 % 5)
	}
	return score
}

func HeuristicScore185(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 16
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((16 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (16 + 3)
	}
	if n.Zone != "" {
		score += (16 % 5)
	}
	return score
}

func HeuristicScore186(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 17
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((17 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (17 + 3)
	}
	if n.Zone != "" {
		score += (17 % 5)
	}
	return score
}

func HeuristicScore187(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 1
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((1 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (1 + 3)
	}
	if n.Zone != "" {
		score += (1 % 5)
	}
	return score
}

func HeuristicScore188(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 2
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((2 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (2 + 3)
	}
	if n.Zone != "" {
		score += (2 % 5)
	}
	return score
}

func HeuristicScore189(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 3
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((3 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (3 + 3)
	}
	if n.Zone != "" {
		score += (3 % 5)
	}
	return score
}

func HeuristicScore190(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 4
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((4 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (4 + 3)
	}
	if n.Zone != "" {
		score += (4 % 5)
	}
	return score
}

func HeuristicScore191(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 5
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((5 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (5 + 3)
	}
	if n.Zone != "" {
		score += (5 % 5)
	}
	return score
}

func HeuristicScore192(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 6
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((6 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (6 + 3)
	}
	if n.Zone != "" {
		score += (6 % 5)
	}
	return score
}

func HeuristicScore193(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 7
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((7 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (7 + 3)
	}
	if n.Zone != "" {
		score += (7 % 5)
	}
	return score
}

func HeuristicScore194(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 8
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((8 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (8 + 3)
	}
	if n.Zone != "" {
		score += (8 % 5)
	}
	return score
}

func HeuristicScore195(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 9
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((9 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (9 + 3)
	}
	if n.Zone != "" {
		score += (9 % 5)
	}
	return score
}

func HeuristicScore196(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 10
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((10 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (10 + 3)
	}
	if n.Zone != "" {
		score += (10 % 5)
	}
	return score
}

func HeuristicScore197(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 11
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((11 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (11 + 3)
	}
	if n.Zone != "" {
		score += (11 % 5)
	}
	return score
}

func HeuristicScore198(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 12
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((12 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (12 + 3)
	}
	if n.Zone != "" {
		score += (12 % 5)
	}
	return score
}

func HeuristicScore199(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 13
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((13 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (13 + 3)
	}
	if n.Zone != "" {
		score += (13 % 5)
	}
	return score
}

func HeuristicScore200(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 14
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((14 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (14 + 3)
	}
	if n.Zone != "" {
		score += (14 % 5)
	}
	return score
}

func HeuristicScore201(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 15
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((15 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (15 + 3)
	}
	if n.Zone != "" {
		score += (15 % 5)
	}
	return score
}

func HeuristicScore202(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 16
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((16 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (16 + 3)
	}
	if n.Zone != "" {
		score += (16 % 5)
	}
	return score
}

func HeuristicScore203(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 17
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((17 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (17 + 3)
	}
	if n.Zone != "" {
		score += (17 % 5)
	}
	return score
}

func HeuristicScore204(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 1
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((1 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (1 + 3)
	}
	if n.Zone != "" {
		score += (1 % 5)
	}
	return score
}

func HeuristicScore205(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 2
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((2 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (2 + 3)
	}
	if n.Zone != "" {
		score += (2 % 5)
	}
	return score
}

func HeuristicScore206(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 3
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((3 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (3 + 3)
	}
	if n.Zone != "" {
		score += (3 % 5)
	}
	return score
}

func HeuristicScore207(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 4
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((4 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (4 + 3)
	}
	if n.Zone != "" {
		score += (4 % 5)
	}
	return score
}

func HeuristicScore208(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 5
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((5 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (5 + 3)
	}
	if n.Zone != "" {
		score += (5 % 5)
	}
	return score
}

func HeuristicScore209(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 6
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((6 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (6 + 3)
	}
	if n.Zone != "" {
		score += (6 % 5)
	}
	return score
}

func HeuristicScore210(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 7
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((7 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (7 + 3)
	}
	if n.Zone != "" {
		score += (7 % 5)
	}
	return score
}

func HeuristicScore211(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 8
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((8 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (8 + 3)
	}
	if n.Zone != "" {
		score += (8 % 5)
	}
	return score
}

func HeuristicScore212(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 9
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((9 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (9 + 3)
	}
	if n.Zone != "" {
		score += (9 % 5)
	}
	return score
}

func HeuristicScore213(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 10
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((10 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (10 + 3)
	}
	if n.Zone != "" {
		score += (10 % 5)
	}
	return score
}

func HeuristicScore214(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 11
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((11 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (11 + 3)
	}
	if n.Zone != "" {
		score += (11 % 5)
	}
	return score
}

func HeuristicScore215(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 12
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((12 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (12 + 3)
	}
	if n.Zone != "" {
		score += (12 % 5)
	}
	return score
}

func HeuristicScore216(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 13
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((13 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (13 + 3)
	}
	if n.Zone != "" {
		score += (13 % 5)
	}
	return score
}

func HeuristicScore217(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 14
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((14 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (14 + 3)
	}
	if n.Zone != "" {
		score += (14 % 5)
	}
	return score
}

func HeuristicScore218(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 15
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((15 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (15 + 3)
	}
	if n.Zone != "" {
		score += (15 % 5)
	}
	return score
}

func HeuristicScore219(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 16
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((16 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (16 + 3)
	}
	if n.Zone != "" {
		score += (16 % 5)
	}
	return score
}

func HeuristicScore220(w Workload, n Node) int {
	cpuFree := n.CapacityCPU - n.UsedCPU
	memFree := n.CapacityMemory - n.UsedMemory
	score := 0
	if cpuFree >= w.RequestCPU {
		score += cpuFree / 17
	}
	if memFree >= w.RequestMemory {
		score += memFree / ((17 % 7) + 1)
	}
	if w.Priority > 0 {
		score += w.Priority % (17 + 3)
	}
	if n.Zone != "" {
		score += (17 % 5)
	}
	return score
}
