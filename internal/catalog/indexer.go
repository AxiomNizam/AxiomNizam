package catalog

// =====================================================
// WS-1.1 — Catalog Full-Text Search Index
//
// Provides in-memory full-text search over catalog asset metadata.
// Indexes asset names, descriptions, tags, domains, owners, and
// column names for fast discovery queries.
//
// In production, this could be backed by Bleve or Tantivy for
// persistence and advanced ranking. The current implementation
// uses an inverted index in memory, rebuilt on startup from etcd.
// =====================================================

import (
	"sort"
	"strings"
	"sync"
)

// SearchResult represents a single search hit.
type SearchResult struct {
	AssetName   string  `json:"assetName"`
	AssetType   string  `json:"assetType"`
	Domain      string  `json:"domain"`
	Description string  `json:"description"`
	Score       float64 `json:"score"`
	Highlights  []string `json:"highlights,omitempty"`
}

// SearchOptions controls search behavior.
type SearchOptions struct {
	Query      string   `json:"query"`
	Domain     string   `json:"domain,omitempty"`
	AssetType  string   `json:"assetType,omitempty"`
	Owner      string   `json:"owner,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	MaxResults int      `json:"maxResults,omitempty"`
	Offset     int      `json:"offset,omitempty"`
}

// SearchResponse wraps search results with pagination metadata.
type SearchResponse struct {
	Results    []SearchResult `json:"results"`
	TotalCount int            `json:"totalCount"`
	Query      string         `json:"query"`
	Took       string         `json:"took"`
}

// CatalogIndexer provides full-text search over catalog assets.
type CatalogIndexer struct {
	mu       sync.RWMutex
	index    map[string][]indexEntry // token -> entries
	assets   map[string]*indexedAsset
}

// indexEntry maps a token to an asset with a relevance weight.
type indexEntry struct {
	AssetName string
	Field     string  // which field matched (name, description, tag, column, etc.)
	Weight    float64 // relevance weight for this field
}

// indexedAsset stores the indexed metadata for an asset.
type indexedAsset struct {
	Name        string
	AssetType   string
	Domain      string
	Owner       string
	Description string
	Tags        []string
	Columns     []string
}

// NewCatalogIndexer creates a new empty indexer.
func NewCatalogIndexer() *CatalogIndexer {
	return &CatalogIndexer{
		index:  make(map[string][]indexEntry),
		assets: make(map[string]*indexedAsset),
	}
}

// IndexAsset adds or updates an asset in the search index.
func (idx *CatalogIndexer) IndexAsset(asset *CatalogAssetResource) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	name := asset.Name

	// Remove old entries if re-indexing.
	idx.removeAssetLocked(name)

	// Build indexed asset.
	ia := &indexedAsset{
		Name:        name,
		AssetType:   string(asset.Spec.AssetType),
		Domain:      asset.Spec.Domain,
		Owner:       asset.Spec.Owner,
		Description: asset.Spec.Description,
		Tags:        asset.Spec.Tags,
	}
	for _, col := range asset.Spec.Columns {
		ia.Columns = append(ia.Columns, col.Name)
	}
	idx.assets[name] = ia

	// Index name tokens (highest weight).
	for _, token := range tokenize(name) {
		idx.index[token] = append(idx.index[token], indexEntry{AssetName: name, Field: "name", Weight: 10.0})
	}

	// Index description tokens.
	for _, token := range tokenize(asset.Spec.Description) {
		idx.index[token] = append(idx.index[token], indexEntry{AssetName: name, Field: "description", Weight: 3.0})
	}

	// Index domain.
	if asset.Spec.Domain != "" {
		for _, token := range tokenize(asset.Spec.Domain) {
			idx.index[token] = append(idx.index[token], indexEntry{AssetName: name, Field: "domain", Weight: 5.0})
		}
	}

	// Index owner.
	if asset.Spec.Owner != "" {
		for _, token := range tokenize(asset.Spec.Owner) {
			idx.index[token] = append(idx.index[token], indexEntry{AssetName: name, Field: "owner", Weight: 4.0})
		}
	}

	// Index tags (high weight — explicit categorization).
	for _, tag := range asset.Spec.Tags {
		for _, token := range tokenize(tag) {
			idx.index[token] = append(idx.index[token], indexEntry{AssetName: name, Field: "tag", Weight: 7.0})
		}
	}

	// Index column names.
	for _, col := range asset.Spec.Columns {
		for _, token := range tokenize(col.Name) {
			idx.index[token] = append(idx.index[token], indexEntry{AssetName: name, Field: "column", Weight: 2.0})
		}
	}

	// Index table name.
	if asset.Spec.TableName != "" {
		for _, token := range tokenize(asset.Spec.TableName) {
			idx.index[token] = append(idx.index[token], indexEntry{AssetName: name, Field: "tableName", Weight: 8.0})
		}
	}

	// Index database and schema.
	if asset.Spec.Database != "" {
		for _, token := range tokenize(asset.Spec.Database) {
			idx.index[token] = append(idx.index[token], indexEntry{AssetName: name, Field: "database", Weight: 4.0})
		}
	}
}

// RemoveAsset removes an asset from the index.
func (idx *CatalogIndexer) RemoveAsset(name string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.removeAssetLocked(name)
}

// removeAssetLocked removes an asset without acquiring the lock.
func (idx *CatalogIndexer) removeAssetLocked(name string) {
	delete(idx.assets, name)

	// Remove all index entries for this asset.
	for token, entries := range idx.index {
		filtered := entries[:0]
		for _, e := range entries {
			if e.AssetName != name {
				filtered = append(filtered, e)
			}
		}
		if len(filtered) == 0 {
			delete(idx.index, token)
		} else {
			idx.index[token] = filtered
		}
	}
}

// Search performs a full-text search with optional facet filtering.
func (idx *CatalogIndexer) Search(opts SearchOptions) *SearchResponse {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if opts.MaxResults <= 0 {
		opts.MaxResults = 50
	}

	// Tokenize query.
	queryTokens := tokenize(opts.Query)
	if len(queryTokens) == 0 && opts.Domain == "" && opts.AssetType == "" && opts.Owner == "" && len(opts.Tags) == 0 {
		return &SearchResponse{Query: opts.Query}
	}

	// Score each asset.
	scores := make(map[string]float64)
	highlights := make(map[string][]string)

	for _, token := range queryTokens {
		// Exact match.
		if entries, ok := idx.index[token]; ok {
			for _, e := range entries {
				scores[e.AssetName] += e.Weight
				highlights[e.AssetName] = appendUnique(highlights[e.AssetName], e.Field+":"+token)
			}
		}

		// Prefix match (for partial queries).
		for indexToken, entries := range idx.index {
			if indexToken != token && strings.HasPrefix(indexToken, token) && len(token) >= 2 {
				for _, e := range entries {
					scores[e.AssetName] += e.Weight * 0.5 // Partial match gets half weight
					highlights[e.AssetName] = appendUnique(highlights[e.AssetName], e.Field+":"+indexToken)
				}
			}
		}
	}

	// If no query tokens but filters exist, score all assets equally.
	if len(queryTokens) == 0 {
		for name := range idx.assets {
			scores[name] = 1.0
		}
	}

	// Apply facet filters.
	var results []SearchResult
	for name, score := range scores {
		asset, ok := idx.assets[name]
		if !ok {
			continue
		}

		// Filter by domain.
		if opts.Domain != "" && !strings.EqualFold(asset.Domain, opts.Domain) {
			continue
		}

		// Filter by asset type.
		if opts.AssetType != "" && !strings.EqualFold(asset.AssetType, opts.AssetType) {
			continue
		}

		// Filter by owner.
		if opts.Owner != "" && !strings.EqualFold(asset.Owner, opts.Owner) {
			continue
		}

		// Filter by tags (all specified tags must be present).
		if len(opts.Tags) > 0 && !hasAllTags(asset.Tags, opts.Tags) {
			continue
		}

		results = append(results, SearchResult{
			AssetName:   name,
			AssetType:   asset.AssetType,
			Domain:      asset.Domain,
			Description: asset.Description,
			Score:       score,
			Highlights:  highlights[name],
		})
	}

	// Sort by score descending.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	totalCount := len(results)

	// Apply pagination.
	if opts.Offset > 0 && opts.Offset < len(results) {
		results = results[opts.Offset:]
	} else if opts.Offset >= len(results) {
		results = nil
	}
	if len(results) > opts.MaxResults {
		results = results[:opts.MaxResults]
	}

	return &SearchResponse{
		Results:    results,
		TotalCount: totalCount,
		Query:      opts.Query,
	}
}

// AssetCount returns the number of indexed assets.
func (idx *CatalogIndexer) AssetCount() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.assets)
}

// TokenCount returns the number of unique tokens in the index.
func (idx *CatalogIndexer) TokenCount() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.index)
}

// --- Helper functions ---

// tokenize splits text into lowercase tokens for indexing.
func tokenize(text string) []string {
	if text == "" {
		return nil
	}

	// Normalize: lowercase, replace separators with spaces.
	normalized := strings.ToLower(text)
	normalized = strings.NewReplacer(
		"_", " ",
		"-", " ",
		".", " ",
		"/", " ",
		"\\", " ",
		"@", " ",
		":", " ",
	).Replace(normalized)

	// Split on whitespace.
	parts := strings.Fields(normalized)

	// Filter out very short tokens and stop words.
	var tokens []string
	for _, p := range parts {
		if len(p) < 2 {
			continue
		}
		if isStopWord(p) {
			continue
		}
		tokens = append(tokens, p)
	}

	return tokens
}

// isStopWord returns true for common words that shouldn't be indexed.
func isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "and": true, "for": true, "are": true,
		"but": true, "not": true, "you": true, "all": true,
		"can": true, "had": true, "her": true, "was": true,
		"one": true, "our": true, "out": true, "has": true,
		"with": true, "this": true, "that": true, "from": true,
		"they": true, "been": true, "have": true, "its": true,
		"will": true, "each": true, "make": true, "like": true,
	}
	return stopWords[word]
}

// hasAllTags checks if the asset has all required tags.
func hasAllTags(assetTags, requiredTags []string) bool {
	tagSet := make(map[string]bool)
	for _, t := range assetTags {
		tagSet[strings.ToLower(t)] = true
	}
	for _, req := range requiredTags {
		if !tagSet[strings.ToLower(req)] {
			return false
		}
	}
	return true
}

// appendUnique appends a string to a slice only if not already present.
func appendUnique(slice []string, s string) []string {
	for _, existing := range slice {
		if existing == s {
			return slice
		}
	}
	return append(slice, s)
}
