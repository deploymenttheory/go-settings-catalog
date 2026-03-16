package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	GroupCollectionType = "#microsoft.graph.deviceManagementConfigurationSettingGroupCollectionDefinition"
)

// Catalog manages the Intune Settings Catalog data
type Catalog struct {
	entriesById       map[string]*CatalogEntry
	categoriesById    map[string]*CatalogCategory
	groupsByOffsetURI map[string][]*CatalogEntry
	catalogDate       string
}

// NewCatalog creates a new empty catalog
func NewCatalog() *Catalog {
	return &Catalog{
		entriesById:       make(map[string]*CatalogEntry),
		categoriesById:    make(map[string]*CatalogCategory),
		groupsByOffsetURI: make(map[string][]*CatalogEntry),
	}
}

// LoadFromBytes loads catalog data from byte slices (for embedded data)
func (c *Catalog) LoadFromBytes(catalogData, categoriesData, versionData []byte) error {
	// Parse settings catalog
	var rawEntries []RawCatalogEntry
	if err := json.Unmarshal(catalogData, &rawEntries); err != nil {
		return fmt.Errorf("failed to parse catalog JSON: %w", err)
	}

	// Parse categories
	var categories []CatalogCategory
	if err := json.Unmarshal(categoriesData, &categories); err != nil {
		return fmt.Errorf("failed to parse categories JSON: %w", err)
	}

	// Parse version
	var versionInfo map[string]string
	if err := json.Unmarshal(versionData, &versionInfo); err != nil {
		return fmt.Errorf("failed to parse version JSON: %w", err)
	}

	c.catalogDate = versionInfo["date"]
	c.loadEntries(rawEntries)
	c.loadCategories(categories)

	return nil
}

// LoadFromFiles loads catalog data from JSON files
func (c *Catalog) LoadFromFiles(catalogPath, categoriesPath, versionPath string) error {
	catalogData, err := os.ReadFile(catalogPath)
	if err != nil {
		return fmt.Errorf("failed to read catalog file: %w", err)
	}

	categoriesData, err := os.ReadFile(categoriesPath)
	if err != nil {
		return fmt.Errorf("failed to read categories file: %w", err)
	}

	versionData, err := os.ReadFile(versionPath)
	if err != nil {
		return fmt.Errorf("failed to read version file: %w", err)
	}

	return c.LoadFromBytes(catalogData, categoriesData, versionData)
}

func (c *Catalog) loadEntries(rawEntries []RawCatalogEntry) {
	for _, raw := range rawEntries {
		if raw.ID == "" || raw.OffsetURI == "" || raw.ODataType == "" {
			continue
		}

		// Parse options
		options := make([]CatalogOption, 0, len(raw.Options))
		for _, opt := range raw.Options {
			itemID, _ := opt["itemId"].(string)
			if itemID == "" {
				continue
			}

			optionValue, _ := opt["optionValue"].(map[string]any)
			var optVal string

			// Handle both string and int option values
			if val, ok := optionValue["value"].(string); ok {
				optVal = val
			} else if val, ok := optionValue["value"].(float64); ok {
				optVal = fmt.Sprintf("%.0f", val)
			} else {
				continue
			}

			displayName, _ := opt["displayName"].(string)
			if displayName == "" {
				displayName = optVal
			}

			options = append(options, CatalogOption{
				ItemID:      itemID,
				OptionValue: optVal,
				DisplayName: displayName,
			})
		}

		entry := &CatalogEntry{
			ID:              raw.ID,
			OffsetURI:       raw.OffsetURI,
			ODataType:       raw.ODataType,
			Options:         options,
			ChildIDs:        raw.ChildIDs,
			DisplayName:     raw.DisplayName,
			Description:     raw.Description,
			Keywords:        raw.Keywords,
			InfoURLs:        raw.InfoURLs,
			CategoryID:      raw.CategoryID,
			ValueDefinition: raw.ValueDefinition,
		}

		if entry.DisplayName == "" {
			entry.DisplayName = entry.OffsetURI
		}

		c.entriesById[entry.ID] = entry

		// Index group collection entries by offsetUri
		if entry.ODataType == GroupCollectionType {
			key := strings.ToLower(entry.OffsetURI)
			c.groupsByOffsetURI[key] = append(c.groupsByOffsetURI[key], entry)
		}
	}
}

func (c *Catalog) loadCategories(categories []CatalogCategory) {
	for _, cat := range categories {
		if cat.ID == "" {
			continue
		}
		c.categoriesById[cat.ID] = &cat
	}
}

// RootGroups returns all root-group entries whose offsetUri matches the given PayloadType
func (c *Catalog) RootGroups(payloadType string) []*CatalogEntry {
	key := strings.ToLower(payloadType)
	return c.groupsByOffsetURI[key]
}

// Entry returns the entry for the given catalog ID
func (c *Catalog) Entry(id string) *CatalogEntry {
	return c.entriesById[id]
}

// MatchResult contains the result of a key match operation
type MatchResult struct {
	Entry           *CatalogEntry
	MatchType       string
	SimilarityScore float64
	NearestMatches  []NearestMatch
}

// NearestMatch represents a potential catalog match
type NearestMatch struct {
	Entry           *CatalogEntry
	SimilarityScore float64
}

// FindChild finds the child entry in childIds whose offsetUri last path component matches key
// Uses fuzzy matching with a similarity threshold of 0.75
func (c *Catalog) FindChild(key string, childIds []string) *CatalogEntry {
	result := c.FindChildWithDetails(key, childIds)
	return result.Entry
}

// FindChildWithDetails finds the child entry and returns detailed match information
func (c *Catalog) FindChildWithDetails(key string, childIds []string) MatchResult {
	keyLower := strings.ToLower(key)
	const similarityThreshold = 0.75
	const nearMatchCount = 3
	
	// Collect all candidates with scores
	type candidate struct {
		entry *CatalogEntry
		score float64
	}
	var candidates []candidate
	
	for _, childID := range childIds {
		child := c.entriesById[childID]
		if child == nil {
			continue
		}
		
		catalogKey := strings.ToLower(offsetKey(child.OffsetURI))
		
		// Check for exact match
		if catalogKey == keyLower {
			return MatchResult{
				Entry:           child,
				MatchType:       "exact",
				SimilarityScore: 1.0,
			}
		}
		
		// Calculate fuzzy score
		score := stringSimilarity(keyLower, catalogKey)
		candidates = append(candidates, candidate{entry: child, score: score})
	}
	
	// Sort candidates by score (descending)
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].score > candidates[i].score {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}
	
	// Check if best candidate meets threshold
	if len(candidates) > 0 && candidates[0].score >= similarityThreshold {
		// Collect near matches for reporting
		var nearMatches []NearestMatch
		for i := 0; i < len(candidates) && i < nearMatchCount; i++ {
			nearMatches = append(nearMatches, NearestMatch{
				Entry:           candidates[i].entry,
				SimilarityScore: candidates[i].score,
			})
		}
		
		return MatchResult{
			Entry:           candidates[0].entry,
			MatchType:       "fuzzy",
			SimilarityScore: candidates[0].score,
			NearestMatches:  nearMatches,
		}
	}
	
	// No match found - return top candidates as near matches
	var nearMatches []NearestMatch
	for i := 0; i < len(candidates) && i < nearMatchCount; i++ {
		nearMatches = append(nearMatches, NearestMatch{
			Entry:           candidates[i].entry,
			SimilarityScore: candidates[i].score,
		})
	}
	
	return MatchResult{
		Entry:          nil,
		MatchType:      "none",
		NearestMatches: nearMatches,
	}
}

// stringSimilarity calculates similarity between two strings using Levenshtein distance
// Returns a score between 0.0 (completely different) and 1.0 (identical)
func stringSimilarity(s1, s2 string) float64 {
	// Normalize strings by removing hyphens and underscores for comparison
	norm1 := strings.ReplaceAll(strings.ReplaceAll(s1, "-", ""), "_", "")
	norm2 := strings.ReplaceAll(strings.ReplaceAll(s2, "-", ""), "_", "")
	
	if norm1 == norm2 {
		return 1.0
	}
	
	// Calculate Levenshtein distance
	distance := levenshteinDistance(norm1, norm2)
	maxLen := max(len(norm1), len(norm2))
	
	if maxLen == 0 {
		return 0.0
	}
	
	return 1.0 - float64(distance)/float64(maxLen)
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}
	
	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}
	
	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}
	
	return matrix[len(s1)][len(s2)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// CatalogDate returns the human-readable catalog date
func (c *Catalog) CatalogDate() string {
	return c.catalogDate
}

// offsetKey extracts the key portion of an offsetUri by taking the last non-index path component
// "Applications/[{0}]/BundleID" → "BundleID"
// "EnableFirewall" → "EnableFirewall"
func offsetKey(uri string) string {
	parts := strings.Split(uri, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if !strings.HasPrefix(parts[i], "[") {
			return parts[i]
		}
	}
	return uri
}
