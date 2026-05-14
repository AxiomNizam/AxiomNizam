package anonymization

// =====================================================
// WS-7.3 — Synthetic Data Generator
//
// Generates realistic synthetic datasets that preserve
// statistical properties of the original data while
// eliminating all PII. Uses distribution-aware generation
// for numeric, categorical, and temporal fields.
// =====================================================

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

// SyntheticGenerator produces realistic synthetic data.
type SyntheticGenerator struct {
	rng    *rand.Rand
	locale string
}

// NewSyntheticGenerator creates a new synthetic data generator.
func NewSyntheticGenerator(seed int64) *SyntheticGenerator {
	return &SyntheticGenerator{
		rng:    rand.New(rand.NewSource(seed)),
		locale: "en_US",
	}
}

// ColumnProfile describes the statistical profile of a column for synthetic generation.
type ColumnProfile struct {
	Name         string    `json:"name"`
	Type         string    `json:"type"` // string, int, float, date, email, phone, name, address, id
	Min          float64   `json:"min,omitempty"`
	Max          float64   `json:"max,omitempty"`
	Mean         float64   `json:"mean,omitempty"`
	StdDev       float64   `json:"stddev,omitempty"`
	NullRatio    float64   `json:"nullRatio,omitempty"` // 0.0 - 1.0
	UniqueRatio  float64   `json:"uniqueRatio,omitempty"`
	Categories   []string  `json:"categories,omitempty"` // For categorical fields
	DateMin      time.Time `json:"dateMin,omitempty"`
	DateMax      time.Time `json:"dateMax,omitempty"`
	Pattern      string    `json:"pattern,omitempty"` // Regex-like pattern hint
}

// SyntheticDataset represents a generated synthetic dataset.
type SyntheticDataset struct {
	Columns     []string                 `json:"columns"`
	Rows        []map[string]interface{} `json:"rows"`
	RowCount    int                      `json:"rowCount"`
	GeneratedAt time.Time               `json:"generatedAt"`
	Seed        int64                    `json:"seed"`
}

// Generate creates a synthetic dataset based on column profiles.
func (g *SyntheticGenerator) Generate(profiles []ColumnProfile, rowCount int) *SyntheticDataset {
	dataset := &SyntheticDataset{
		RowCount:    rowCount,
		GeneratedAt: time.Now(),
	}

	for _, p := range profiles {
		dataset.Columns = append(dataset.Columns, p.Name)
	}

	for i := 0; i < rowCount; i++ {
		row := make(map[string]interface{})
		for _, profile := range profiles {
			// Handle nulls.
			if profile.NullRatio > 0 && g.rng.Float64() < profile.NullRatio {
				row[profile.Name] = nil
				continue
			}
			row[profile.Name] = g.generateValue(profile, i)
		}
		dataset.Rows = append(dataset.Rows, row)
	}

	return dataset
}

// generateValue creates a single synthetic value based on the column profile.
func (g *SyntheticGenerator) generateValue(profile ColumnProfile, index int) interface{} {
	switch profile.Type {
	case "int":
		return g.generateInt(profile)
	case "float":
		return g.generateFloat(profile)
	case "string":
		return g.generateString(profile, index)
	case "date":
		return g.generateDate(profile)
	case "email":
		return g.generateEmail(index)
	case "phone":
		return g.generatePhone()
	case "name":
		return g.generateName()
	case "address":
		return g.generateAddress()
	case "id":
		return g.generateID(profile, index)
	case "bool":
		return g.rng.Float64() > 0.5
	case "categorical":
		return g.generateCategorical(profile)
	default:
		return g.generateString(profile, index)
	}
}

func (g *SyntheticGenerator) generateInt(p ColumnProfile) int64 {
	if p.Mean != 0 && p.StdDev != 0 {
		// Normal distribution.
		return int64(g.normalDist(p.Mean, p.StdDev))
	}
	min := int64(p.Min)
	max := int64(p.Max)
	if max <= min {
		max = min + 1000
	}
	return min + g.rng.Int63n(max-min+1)
}

func (g *SyntheticGenerator) generateFloat(p ColumnProfile) float64 {
	if p.Mean != 0 && p.StdDev != 0 {
		return g.normalDist(p.Mean, p.StdDev)
	}
	min := p.Min
	max := p.Max
	if max <= min {
		max = min + 1000.0
	}
	return min + g.rng.Float64()*(max-min)
}

func (g *SyntheticGenerator) generateString(p ColumnProfile, index int) string {
	if len(p.Categories) > 0 {
		return p.Categories[g.rng.Intn(len(p.Categories))]
	}
	return fmt.Sprintf("val_%d_%d", index, g.rng.Intn(10000))
}

func (g *SyntheticGenerator) generateDate(p ColumnProfile) string {
	min := p.DateMin
	max := p.DateMax
	if min.IsZero() {
		min = time.Now().AddDate(-2, 0, 0)
	}
	if max.IsZero() {
		max = time.Now()
	}
	diff := max.Sub(min)
	offset := time.Duration(g.rng.Int63n(int64(diff)))
	return min.Add(offset).Format("2006-01-02")
}

func (g *SyntheticGenerator) generateEmail(index int) string {
	first := firstNames[g.rng.Intn(len(firstNames))]
	last := lastNames[g.rng.Intn(len(lastNames))]
	domains := []string{"example.com", "test.org", "sample.net", "demo.io", "synth.dev"}
	return fmt.Sprintf("%s.%s%d@%s",
		strings.ToLower(first),
		strings.ToLower(last),
		g.rng.Intn(100),
		domains[g.rng.Intn(len(domains))])
}

func (g *SyntheticGenerator) generatePhone() string {
	return fmt.Sprintf("+1-%03d-%03d-%04d",
		g.rng.Intn(900)+100,
		g.rng.Intn(900)+100,
		g.rng.Intn(9000)+1000)
}

func (g *SyntheticGenerator) generateName() string {
	return firstNames[g.rng.Intn(len(firstNames))] + " " + lastNames[g.rng.Intn(len(lastNames))]
}

func (g *SyntheticGenerator) generateAddress() string {
	streets := []string{"Main St", "Oak Ave", "Pine Rd", "Elm Dr", "Cedar Ln", "Maple Ct", "Park Blvd", "Lake Way"}
	cities := []string{"Springfield", "Portland", "Madison", "Franklin", "Georgetown", "Bristol", "Fairview", "Salem"}
	states := []string{"CA", "NY", "TX", "FL", "IL", "PA", "OH", "GA"}

	return fmt.Sprintf("%d %s, %s, %s %05d",
		g.rng.Intn(9999)+1,
		streets[g.rng.Intn(len(streets))],
		cities[g.rng.Intn(len(cities))],
		states[g.rng.Intn(len(states))],
		g.rng.Intn(90000)+10000)
}

func (g *SyntheticGenerator) generateID(p ColumnProfile, index int) string {
	if p.Pattern != "" {
		return fmt.Sprintf("%s-%06d", p.Pattern, index+1)
	}
	return fmt.Sprintf("ID-%08d", g.rng.Intn(99999999))
}

func (g *SyntheticGenerator) generateCategorical(p ColumnProfile) string {
	if len(p.Categories) == 0 {
		return "unknown"
	}
	return p.Categories[g.rng.Intn(len(p.Categories))]
}

func (g *SyntheticGenerator) normalDist(mean, stddev float64) float64 {
	// Box-Muller transform for normal distribution.
	u1 := g.rng.Float64()
	u2 := g.rng.Float64()
	z := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
	return mean + z*stddev
}

// --- Reference data ---

var firstNames = []string{
	"Alice", "Bob", "Carol", "Dave", "Eve", "Frank", "Grace", "Heidi",
	"Ivan", "Judy", "Karl", "Lisa", "Mike", "Nina", "Oscar", "Pat",
	"Quinn", "Ruth", "Sam", "Tina", "Uma", "Victor", "Wendy", "Xavier",
}

var lastNames = []string{
	"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller",
	"Davis", "Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez",
	"Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin",
}
