package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	mustWrite(filepath.Join(root, "internal", "kubeplus", "admission", "policies_generated.go"), genAdmission())
	mustWrite(filepath.Join(root, "internal", "kubeplus", "scheduler", "heuristics_generated.go"), genScheduler())
	mustWrite(filepath.Join(root, "internal", "kubeplus", "crd", "validators_generated.go"), genCRD())
	mustWrite(filepath.Join(root, "internal", "netintel", "modes", "detectors_generated.go"), genNetintelModes())
	mustWrite(filepath.Join(root, "internal", "vectorplus", "similarity_generated.go"), genVector())
	mustWrite(filepath.Join(root, "internal", "reviewflow", "quality_generated.go"), genReviewflow())

	fmt.Println("generated feature module files")
}

func mustWrite(path, content string) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		panic(err)
	}
}

func genAdmission() string {
	var b strings.Builder
	b.WriteString("package admission\n\n")
	b.WriteString("import (\n\t\"fmt\"\n\t\"strings\"\n)\n\n")
	for i := 1; i <= 220; i++ {
		n := fmt.Sprintf("%03d", i)
		b.WriteString(fmt.Sprintf("func PolicyTemplate%s(req AdmissionRequest) AdmissionDecision {\n", n))
		b.WriteString("\tkey := strings.ToLower(req.Kind + \"-\" + req.Operation)\n")
		b.WriteString(fmt.Sprintf("\treason := fmt.Sprintf(\"policy-%s evaluated for %%s\", key)\n", n))
		b.WriteString(fmt.Sprintf("\tif strings.Contains(key, \"deny-%s\") {\n", n))
		b.WriteString(fmt.Sprintf("\t\treturn AdmissionDecision{Allowed: false, Severity: \"high\", Reason: reason, Mutations: map[string]string{\"policy\": \"%s\"}}\n", n))
		b.WriteString("\t}\n")
		b.WriteString(fmt.Sprintf("\tout := map[string]string{\"policy\": \"%s\", \"checked\": \"true\"}\n", n))
		b.WriteString("\tif req.Labels != nil {\n")
		b.WriteString("\t\tif v, ok := req.Labels[\"mode\"]; ok && strings.TrimSpace(v) != \"\" {\n")
		b.WriteString("\t\t\tout[\"mode\"] = v\n")
		b.WriteString("\t\t}\n")
		b.WriteString("\t}\n")
		b.WriteString("\treturn AdmissionDecision{Allowed: true, Severity: \"info\", Reason: reason, Mutations: out}\n")
		b.WriteString("}\n\n")
	}
	return b.String()
}

func genScheduler() string {
	var b strings.Builder
	b.WriteString("package scheduler\n\n")
	for i := 1; i <= 220; i++ {
		n := fmt.Sprintf("%03d", i)
		m := (i % 17) + 1
		b.WriteString(fmt.Sprintf("func HeuristicScore%s(w Workload, n Node) int {\n", n))
		b.WriteString("\tcpuFree := n.CapacityCPU - n.UsedCPU\n")
		b.WriteString("\tmemFree := n.CapacityMemory - n.UsedMemory\n")
		b.WriteString("\tscore := 0\n")
		b.WriteString(fmt.Sprintf("\tif cpuFree >= w.RequestCPU { score += cpuFree / %d }\n", m))
		b.WriteString(fmt.Sprintf("\tif memFree >= w.RequestMemory { score += memFree / ((%d %% 7) + 1) }\n", m))
		b.WriteString(fmt.Sprintf("\tif w.Priority > 0 { score += w.Priority %% (%d + 3) }\n", m))
		b.WriteString(fmt.Sprintf("\tif n.Zone != \"\" { score += (%d %% 5) }\n", m))
		b.WriteString("\treturn score\n")
		b.WriteString("}\n\n")
	}
	return b.String()
}

func genCRD() string {
	var b strings.Builder
	b.WriteString("package crd\n\n")
	for i := 1; i <= 220; i++ {
		n := fmt.Sprintf("%03d", i)
		b.WriteString(fmt.Sprintf("func ValidateTemplate%s(spec map[string]any) ValidationResult {\n", n))
		b.WriteString("\tres := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}\n")
		b.WriteString("\tif spec == nil {\n")
		b.WriteString("\t\treturn ValidationResult{Valid: false, Errors: []string{\"spec cannot be nil\"}, Warnings: []string{}}\n")
		b.WriteString("\t}\n")
		b.WriteString("\tif _, ok := spec[\"name\"]; !ok {\n")
		b.WriteString("\t\tres.Valid = false\n")
		b.WriteString(fmt.Sprintf("\t\tres.Errors = append(res.Errors, \"missing name for template %s\")\n", n))
		b.WriteString("\t}\n")
		b.WriteString("\tif v, ok := spec[\"replicas\"]; ok {\n")
		b.WriteString("\t\tswitch v.(type) {\n")
		b.WriteString("\t\tcase int, int32, int64, float64:\n")
		b.WriteString("\t\tdefault:\n")
		b.WriteString("\t\t\tres.Valid = false\n")
		b.WriteString("\t\t\tres.Errors = append(res.Errors, \"replicas must be numeric\")\n")
		b.WriteString("\t\t}\n")
		b.WriteString("\t} else {\n")
		b.WriteString("\t\tres.Warnings = append(res.Warnings, \"replicas not set, defaulting to 1\")\n")
		b.WriteString("\t}\n")
		b.WriteString("\treturn res\n")
		b.WriteString("}\n\n")
	}
	return b.String()
}

func genNetintelModes() string {
	var b strings.Builder
	b.WriteString("package modes\n\n")
	b.WriteString("import \"math\"\n\n")
	for i := 1; i <= 220; i++ {
		n := fmt.Sprintf("%03d", i)
		weight := float64((i%9)+1) / 10.0
		b.WriteString(fmt.Sprintf("func Detector%s(samples []float64) float64 {\n", n))
		b.WriteString("\tif len(samples) == 0 {\n\t\treturn 0\n\t}\n")
		b.WriteString("\tvar sum float64\n")
		b.WriteString("\tfor _, s := range samples {\n\t\tsum += s\n\t}\n")
		b.WriteString("\tmean := sum / float64(len(samples))\n")
		b.WriteString("\tvar variance float64\n")
		b.WriteString("\tfor _, s := range samples {\n\t\td := s - mean\n\t\tvariance += d * d\n\t}\n")
		b.WriteString("\tvariance = variance / float64(len(samples))\n")
		b.WriteString("\tsigma := math.Sqrt(variance)\n")
		b.WriteString(fmt.Sprintf("\tweight := %.3f\n", weight))
		b.WriteString("\treturn mean + sigma*weight\n")
		b.WriteString("}\n\n")
	}
	return b.String()
}

func genVector() string {
	var b strings.Builder
	b.WriteString("package vectorplus\n\n")
	b.WriteString("import \"math\"\n\n")
	for i := 1; i <= 220; i++ {
		n := fmt.Sprintf("%03d", i)
		tweak := float64((i%13)+1) / 1000.0
		b.WriteString(fmt.Sprintf("func SimilarityMetric%s(a, b Vector) float64 {\n", n))
		b.WriteString("\tif len(a) == 0 || len(b) == 0 {\n\t\treturn 0\n\t}\n")
		b.WriteString("\tn := len(a)\n")
		b.WriteString("\tif len(b) < n {\n\t\tn = len(b)\n\t}\n")
		b.WriteString("\tvar dotv, na, nb float64\n")
		b.WriteString("\tfor i := 0; i < n; i++ {\n")
		b.WriteString("\t\tdotv += a[i] * b[i]\n")
		b.WriteString("\t\tna += a[i] * a[i]\n")
		b.WriteString("\t\tnb += b[i] * b[i]\n")
		b.WriteString("\t}\n")
		b.WriteString("\tdenom := math.Sqrt(na) * math.Sqrt(nb)\n")
		b.WriteString("\tif denom == 0 {\n\t\treturn 0\n\t}\n")
		b.WriteString(fmt.Sprintf("\ttweak := %.4f\n", tweak))
		b.WriteString("\treturn (dotv / denom) - tweak\n")
		b.WriteString("}\n\n")
	}
	return b.String()
}

func genReviewflow() string {
	var b strings.Builder
	b.WriteString("package reviewflow\n\n")
	b.WriteString("import \"strings\"\n\n")
	for i := 1; i <= 220; i++ {
		n := fmt.Sprintf("%03d", i)
		baseline := float64((i%11)+1) / 10.0
		b.WriteString(fmt.Sprintf("func QualityCheck%s(item ReviewItem) float64 {\n", n))
		b.WriteString("\tscore := 0.0\n")
		b.WriteString("\tif strings.TrimSpace(item.Title) != \"\" { score += 1.0 }\n")
		b.WriteString("\tif strings.TrimSpace(item.Description) != \"\" { score += 1.0 }\n")
		b.WriteString("\tif len(item.Tags) > 0 { score += float64(len(item.Tags)) * 0.2 }\n")
		b.WriteString(fmt.Sprintf("\tif strings.Contains(strings.ToLower(item.Title), \"%s\") { score += 0.3 }\n", n))
		b.WriteString("\tif item.Stage == StageApproved || item.Stage == StageMerged { score += 0.7 }\n")
		b.WriteString(fmt.Sprintf("\tbaseline := %.3f\n", baseline))
		b.WriteString("\treturn score + baseline\n")
		b.WriteString("}\n\n")
	}
	return b.String()
}
