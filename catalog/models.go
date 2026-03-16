package catalog

// CatalogOption represents a single option value for a setting
type CatalogOption struct {
	ItemID      string `json:"itemId"`
	OptionValue string `json:"optionValue"`
	DisplayName string `json:"displayName"`
}

// CatalogEntry represents a setting definition from the Intune catalog
type CatalogEntry struct {
	ID              string           `json:"id"`
	OffsetURI       string           `json:"offsetUri"`
	ODataType       string           `json:"@odata.type"`
	Options         []CatalogOption  `json:"options"`
	ChildIDs        []string         `json:"childIds"`
	DisplayName     string           `json:"displayName"`
	Description     string           `json:"description"`
	Keywords        []string         `json:"keywords"`
	InfoURLs        []string         `json:"infoUrls"`
	CategoryID      string           `json:"categoryId"`
	ValueDefinition *ValueDefinition `json:"valueDefinition,omitempty"`
	ValueMinimum    *int             `json:"valueMinimum,omitempty"`
	ValueMaximum    *int             `json:"valueMaximum,omitempty"`
}

// CatalogCategory represents a category in the Intune catalog
type CatalogCategory struct {
	ID                  string   `json:"id"`
	DisplayName         string   `json:"displayName"`
	Description         string   `json:"description"`
	CategoryDescription string   `json:"categoryDescription"`
	ChildCategoryIDs    []string `json:"childCategoryIds"`
}

// RawCatalogEntry is used for JSON unmarshaling with flexible value types
type RawCatalogEntry struct {
	ID              string           `json:"id"`
	OffsetURI       string           `json:"offsetUri"`
	ODataType       string           `json:"@odata.type"`
	Options         []map[string]any `json:"options"`
	ChildIDs        []string         `json:"childIds"`
	DisplayName     string           `json:"displayName"`
	Description     string           `json:"description"`
	Keywords        []string         `json:"keywords"`
	InfoURLs        []string         `json:"infoUrls"`
	CategoryID      string           `json:"categoryId"`
	ValueDefinition *ValueDefinition `json:"valueDefinition"`
}

// ValueDefinition represents the value constraints and type for a setting
type ValueDefinition struct {
	ODataType     string `json:"@odata.type"`
	MinimumValue  *int   `json:"minimumValue,omitempty"`
	MaximumValue  *int   `json:"maximumValue,omitempty"`
	IsSecret      bool   `json:"isSecret,omitempty"`
	MinimumLength *int   `json:"minimumLength,omitempty"`
	MaximumLength *int   `json:"maximumLength,omitempty"`
}
