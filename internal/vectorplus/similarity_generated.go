package vectorplus

import "math"

func SimilarityMetric001(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0020
	return (dotv / denom) - tweak
}

func SimilarityMetric002(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0030
	return (dotv / denom) - tweak
}

func SimilarityMetric003(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0040
	return (dotv / denom) - tweak
}

func SimilarityMetric004(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0050
	return (dotv / denom) - tweak
}

func SimilarityMetric005(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0060
	return (dotv / denom) - tweak
}

func SimilarityMetric006(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0070
	return (dotv / denom) - tweak
}

func SimilarityMetric007(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0080
	return (dotv / denom) - tweak
}

func SimilarityMetric008(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0090
	return (dotv / denom) - tweak
}

func SimilarityMetric009(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0100
	return (dotv / denom) - tweak
}

func SimilarityMetric010(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0110
	return (dotv / denom) - tweak
}

func SimilarityMetric011(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0120
	return (dotv / denom) - tweak
}

func SimilarityMetric012(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0130
	return (dotv / denom) - tweak
}

func SimilarityMetric013(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0010
	return (dotv / denom) - tweak
}

func SimilarityMetric014(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0020
	return (dotv / denom) - tweak
}

func SimilarityMetric015(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0030
	return (dotv / denom) - tweak
}

func SimilarityMetric016(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0040
	return (dotv / denom) - tweak
}

func SimilarityMetric017(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0050
	return (dotv / denom) - tweak
}

func SimilarityMetric018(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0060
	return (dotv / denom) - tweak
}

func SimilarityMetric019(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0070
	return (dotv / denom) - tweak
}

func SimilarityMetric020(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0080
	return (dotv / denom) - tweak
}

func SimilarityMetric021(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0090
	return (dotv / denom) - tweak
}

func SimilarityMetric022(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0100
	return (dotv / denom) - tweak
}

func SimilarityMetric023(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0110
	return (dotv / denom) - tweak
}

func SimilarityMetric024(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0120
	return (dotv / denom) - tweak
}

func SimilarityMetric025(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0130
	return (dotv / denom) - tweak
}

func SimilarityMetric026(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0010
	return (dotv / denom) - tweak
}

func SimilarityMetric027(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0020
	return (dotv / denom) - tweak
}

func SimilarityMetric028(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0030
	return (dotv / denom) - tweak
}

func SimilarityMetric029(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0040
	return (dotv / denom) - tweak
}

func SimilarityMetric030(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0050
	return (dotv / denom) - tweak
}

func SimilarityMetric031(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0060
	return (dotv / denom) - tweak
}

func SimilarityMetric032(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0070
	return (dotv / denom) - tweak
}

func SimilarityMetric033(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0080
	return (dotv / denom) - tweak
}

func SimilarityMetric034(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0090
	return (dotv / denom) - tweak
}

func SimilarityMetric035(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0100
	return (dotv / denom) - tweak
}

func SimilarityMetric036(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0110
	return (dotv / denom) - tweak
}

func SimilarityMetric037(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0120
	return (dotv / denom) - tweak
}

func SimilarityMetric038(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0130
	return (dotv / denom) - tweak
}

func SimilarityMetric039(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0010
	return (dotv / denom) - tweak
}

func SimilarityMetric040(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0020
	return (dotv / denom) - tweak
}

func SimilarityMetric041(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0030
	return (dotv / denom) - tweak
}

func SimilarityMetric042(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0040
	return (dotv / denom) - tweak
}

func SimilarityMetric043(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0050
	return (dotv / denom) - tweak
}

func SimilarityMetric044(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0060
	return (dotv / denom) - tweak
}

func SimilarityMetric045(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0070
	return (dotv / denom) - tweak
}

func SimilarityMetric046(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0080
	return (dotv / denom) - tweak
}

func SimilarityMetric047(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0090
	return (dotv / denom) - tweak
}

func SimilarityMetric048(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0100
	return (dotv / denom) - tweak
}

func SimilarityMetric049(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0110
	return (dotv / denom) - tweak
}

func SimilarityMetric050(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0120
	return (dotv / denom) - tweak
}

func SimilarityMetric051(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0130
	return (dotv / denom) - tweak
}

func SimilarityMetric052(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0010
	return (dotv / denom) - tweak
}

func SimilarityMetric053(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0020
	return (dotv / denom) - tweak
}

func SimilarityMetric054(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0030
	return (dotv / denom) - tweak
}

func SimilarityMetric055(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0040
	return (dotv / denom) - tweak
}

func SimilarityMetric056(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0050
	return (dotv / denom) - tweak
}

func SimilarityMetric057(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0060
	return (dotv / denom) - tweak
}

func SimilarityMetric058(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0070
	return (dotv / denom) - tweak
}

func SimilarityMetric059(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0080
	return (dotv / denom) - tweak
}

func SimilarityMetric060(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0090
	return (dotv / denom) - tweak
}

func SimilarityMetric061(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0100
	return (dotv / denom) - tweak
}

func SimilarityMetric062(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0110
	return (dotv / denom) - tweak
}

func SimilarityMetric063(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0120
	return (dotv / denom) - tweak
}

func SimilarityMetric064(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0130
	return (dotv / denom) - tweak
}

func SimilarityMetric065(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0010
	return (dotv / denom) - tweak
}

func SimilarityMetric066(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0020
	return (dotv / denom) - tweak
}

func SimilarityMetric067(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0030
	return (dotv / denom) - tweak
}

func SimilarityMetric068(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0040
	return (dotv / denom) - tweak
}

func SimilarityMetric069(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0050
	return (dotv / denom) - tweak
}

func SimilarityMetric070(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0060
	return (dotv / denom) - tweak
}

func SimilarityMetric071(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0070
	return (dotv / denom) - tweak
}

func SimilarityMetric072(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0080
	return (dotv / denom) - tweak
}

func SimilarityMetric073(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0090
	return (dotv / denom) - tweak
}

func SimilarityMetric074(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0100
	return (dotv / denom) - tweak
}

func SimilarityMetric075(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0110
	return (dotv / denom) - tweak
}

func SimilarityMetric076(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0120
	return (dotv / denom) - tweak
}

func SimilarityMetric077(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0130
	return (dotv / denom) - tweak
}

func SimilarityMetric078(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0010
	return (dotv / denom) - tweak
}

func SimilarityMetric079(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0020
	return (dotv / denom) - tweak
}

func SimilarityMetric080(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0030
	return (dotv / denom) - tweak
}

func SimilarityMetric081(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0040
	return (dotv / denom) - tweak
}

func SimilarityMetric082(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0050
	return (dotv / denom) - tweak
}

func SimilarityMetric083(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0060
	return (dotv / denom) - tweak
}

func SimilarityMetric084(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0070
	return (dotv / denom) - tweak
}

func SimilarityMetric085(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0080
	return (dotv / denom) - tweak
}

func SimilarityMetric086(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0090
	return (dotv / denom) - tweak
}

func SimilarityMetric087(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0100
	return (dotv / denom) - tweak
}

func SimilarityMetric088(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0110
	return (dotv / denom) - tweak
}

func SimilarityMetric089(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0120
	return (dotv / denom) - tweak
}

func SimilarityMetric090(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0130
	return (dotv / denom) - tweak
}

func SimilarityMetric091(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0010
	return (dotv / denom) - tweak
}

func SimilarityMetric092(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0020
	return (dotv / denom) - tweak
}

func SimilarityMetric093(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0030
	return (dotv / denom) - tweak
}

func SimilarityMetric094(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0040
	return (dotv / denom) - tweak
}

func SimilarityMetric095(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0050
	return (dotv / denom) - tweak
}

func SimilarityMetric096(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0060
	return (dotv / denom) - tweak
}

func SimilarityMetric097(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0070
	return (dotv / denom) - tweak
}

func SimilarityMetric098(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0080
	return (dotv / denom) - tweak
}

func SimilarityMetric099(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0090
	return (dotv / denom) - tweak
}

func SimilarityMetric100(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0100
	return (dotv / denom) - tweak
}

func SimilarityMetric101(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0110
	return (dotv / denom) - tweak
}

func SimilarityMetric102(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0120
	return (dotv / denom) - tweak
}

func SimilarityMetric103(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0130
	return (dotv / denom) - tweak
}

func SimilarityMetric104(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0010
	return (dotv / denom) - tweak
}

func SimilarityMetric105(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0020
	return (dotv / denom) - tweak
}

func SimilarityMetric106(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0030
	return (dotv / denom) - tweak
}

func SimilarityMetric107(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0040
	return (dotv / denom) - tweak
}

func SimilarityMetric108(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0050
	return (dotv / denom) - tweak
}

func SimilarityMetric109(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0060
	return (dotv / denom) - tweak
}

func SimilarityMetric110(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0070
	return (dotv / denom) - tweak
}

func SimilarityMetric111(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0080
	return (dotv / denom) - tweak
}

func SimilarityMetric112(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0090
	return (dotv / denom) - tweak
}

func SimilarityMetric113(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0100
	return (dotv / denom) - tweak
}

func SimilarityMetric114(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0110
	return (dotv / denom) - tweak
}

func SimilarityMetric115(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0120
	return (dotv / denom) - tweak
}

func SimilarityMetric116(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0130
	return (dotv / denom) - tweak
}

func SimilarityMetric117(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0010
	return (dotv / denom) - tweak
}

func SimilarityMetric118(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0020
	return (dotv / denom) - tweak
}

func SimilarityMetric119(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0030
	return (dotv / denom) - tweak
}

func SimilarityMetric120(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0040
	return (dotv / denom) - tweak
}

func SimilarityMetric121(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0050
	return (dotv / denom) - tweak
}

func SimilarityMetric122(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0060
	return (dotv / denom) - tweak
}

func SimilarityMetric123(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0070
	return (dotv / denom) - tweak
}

func SimilarityMetric124(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0080
	return (dotv / denom) - tweak
}

func SimilarityMetric125(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0090
	return (dotv / denom) - tweak
}

func SimilarityMetric126(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0100
	return (dotv / denom) - tweak
}

func SimilarityMetric127(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0110
	return (dotv / denom) - tweak
}

func SimilarityMetric128(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0120
	return (dotv / denom) - tweak
}

func SimilarityMetric129(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0130
	return (dotv / denom) - tweak
}

func SimilarityMetric130(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0010
	return (dotv / denom) - tweak
}

func SimilarityMetric131(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0020
	return (dotv / denom) - tweak
}

func SimilarityMetric132(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0030
	return (dotv / denom) - tweak
}

func SimilarityMetric133(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0040
	return (dotv / denom) - tweak
}

func SimilarityMetric134(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0050
	return (dotv / denom) - tweak
}

func SimilarityMetric135(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0060
	return (dotv / denom) - tweak
}

func SimilarityMetric136(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0070
	return (dotv / denom) - tweak
}

func SimilarityMetric137(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0080
	return (dotv / denom) - tweak
}

func SimilarityMetric138(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0090
	return (dotv / denom) - tweak
}

func SimilarityMetric139(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0100
	return (dotv / denom) - tweak
}

func SimilarityMetric140(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0110
	return (dotv / denom) - tweak
}

func SimilarityMetric141(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0120
	return (dotv / denom) - tweak
}

func SimilarityMetric142(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0130
	return (dotv / denom) - tweak
}

func SimilarityMetric143(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0010
	return (dotv / denom) - tweak
}

func SimilarityMetric144(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0020
	return (dotv / denom) - tweak
}

func SimilarityMetric145(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0030
	return (dotv / denom) - tweak
}

func SimilarityMetric146(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0040
	return (dotv / denom) - tweak
}

func SimilarityMetric147(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0050
	return (dotv / denom) - tweak
}

func SimilarityMetric148(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0060
	return (dotv / denom) - tweak
}

func SimilarityMetric149(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0070
	return (dotv / denom) - tweak
}

func SimilarityMetric150(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0080
	return (dotv / denom) - tweak
}

func SimilarityMetric151(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0090
	return (dotv / denom) - tweak
}

func SimilarityMetric152(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0100
	return (dotv / denom) - tweak
}

func SimilarityMetric153(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0110
	return (dotv / denom) - tweak
}

func SimilarityMetric154(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0120
	return (dotv / denom) - tweak
}

func SimilarityMetric155(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0130
	return (dotv / denom) - tweak
}

func SimilarityMetric156(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0010
	return (dotv / denom) - tweak
}

func SimilarityMetric157(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0020
	return (dotv / denom) - tweak
}

func SimilarityMetric158(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0030
	return (dotv / denom) - tweak
}

func SimilarityMetric159(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0040
	return (dotv / denom) - tweak
}

func SimilarityMetric160(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0050
	return (dotv / denom) - tweak
}

func SimilarityMetric161(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0060
	return (dotv / denom) - tweak
}

func SimilarityMetric162(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0070
	return (dotv / denom) - tweak
}

func SimilarityMetric163(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0080
	return (dotv / denom) - tweak
}

func SimilarityMetric164(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0090
	return (dotv / denom) - tweak
}

func SimilarityMetric165(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0100
	return (dotv / denom) - tweak
}

func SimilarityMetric166(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0110
	return (dotv / denom) - tweak
}

func SimilarityMetric167(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0120
	return (dotv / denom) - tweak
}

func SimilarityMetric168(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0130
	return (dotv / denom) - tweak
}

func SimilarityMetric169(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0010
	return (dotv / denom) - tweak
}

func SimilarityMetric170(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0020
	return (dotv / denom) - tweak
}

func SimilarityMetric171(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0030
	return (dotv / denom) - tweak
}

func SimilarityMetric172(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0040
	return (dotv / denom) - tweak
}

func SimilarityMetric173(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0050
	return (dotv / denom) - tweak
}

func SimilarityMetric174(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0060
	return (dotv / denom) - tweak
}

func SimilarityMetric175(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0070
	return (dotv / denom) - tweak
}

func SimilarityMetric176(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0080
	return (dotv / denom) - tweak
}

func SimilarityMetric177(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0090
	return (dotv / denom) - tweak
}

func SimilarityMetric178(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0100
	return (dotv / denom) - tweak
}

func SimilarityMetric179(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0110
	return (dotv / denom) - tweak
}

func SimilarityMetric180(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0120
	return (dotv / denom) - tweak
}

func SimilarityMetric181(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0130
	return (dotv / denom) - tweak
}

func SimilarityMetric182(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0010
	return (dotv / denom) - tweak
}

func SimilarityMetric183(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0020
	return (dotv / denom) - tweak
}

func SimilarityMetric184(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0030
	return (dotv / denom) - tweak
}

func SimilarityMetric185(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0040
	return (dotv / denom) - tweak
}

func SimilarityMetric186(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0050
	return (dotv / denom) - tweak
}

func SimilarityMetric187(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0060
	return (dotv / denom) - tweak
}

func SimilarityMetric188(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0070
	return (dotv / denom) - tweak
}

func SimilarityMetric189(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0080
	return (dotv / denom) - tweak
}

func SimilarityMetric190(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0090
	return (dotv / denom) - tweak
}

func SimilarityMetric191(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0100
	return (dotv / denom) - tweak
}

func SimilarityMetric192(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0110
	return (dotv / denom) - tweak
}

func SimilarityMetric193(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0120
	return (dotv / denom) - tweak
}

func SimilarityMetric194(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0130
	return (dotv / denom) - tweak
}

func SimilarityMetric195(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0010
	return (dotv / denom) - tweak
}

func SimilarityMetric196(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0020
	return (dotv / denom) - tweak
}

func SimilarityMetric197(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0030
	return (dotv / denom) - tweak
}

func SimilarityMetric198(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0040
	return (dotv / denom) - tweak
}

func SimilarityMetric199(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0050
	return (dotv / denom) - tweak
}

func SimilarityMetric200(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0060
	return (dotv / denom) - tweak
}

func SimilarityMetric201(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0070
	return (dotv / denom) - tweak
}

func SimilarityMetric202(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0080
	return (dotv / denom) - tweak
}

func SimilarityMetric203(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0090
	return (dotv / denom) - tweak
}

func SimilarityMetric204(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0100
	return (dotv / denom) - tweak
}

func SimilarityMetric205(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0110
	return (dotv / denom) - tweak
}

func SimilarityMetric206(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0120
	return (dotv / denom) - tweak
}

func SimilarityMetric207(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0130
	return (dotv / denom) - tweak
}

func SimilarityMetric208(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0010
	return (dotv / denom) - tweak
}

func SimilarityMetric209(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0020
	return (dotv / denom) - tweak
}

func SimilarityMetric210(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0030
	return (dotv / denom) - tweak
}

func SimilarityMetric211(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0040
	return (dotv / denom) - tweak
}

func SimilarityMetric212(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0050
	return (dotv / denom) - tweak
}

func SimilarityMetric213(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0060
	return (dotv / denom) - tweak
}

func SimilarityMetric214(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0070
	return (dotv / denom) - tweak
}

func SimilarityMetric215(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0080
	return (dotv / denom) - tweak
}

func SimilarityMetric216(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0090
	return (dotv / denom) - tweak
}

func SimilarityMetric217(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0100
	return (dotv / denom) - tweak
}

func SimilarityMetric218(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0110
	return (dotv / denom) - tweak
}

func SimilarityMetric219(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0120
	return (dotv / denom) - tweak
}

func SimilarityMetric220(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dotv, na, nb float64
	for i := 0; i < n; i++ {
		dotv += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	tweak := 0.0130
	return (dotv / denom) - tweak
}
