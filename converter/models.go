package converter

// MatchType indicates how a key was matched to the catalog
type MatchType string

const (
	MatchTypeExact  MatchType = "exact"
	MatchTypeFuzzy  MatchType = "fuzzy"
	MatchTypeNone   MatchType = "none"
)

// MatchedKey represents a key that was successfully mapped to the catalog
type MatchedKey struct {
	Path           string
	Value          string
	CatalogID      string
	CatalogName    string
	MatchType      MatchType
	SimilarityScore float64
}

// SkippedKey represents a key that couldn't be mapped to the catalog
type SkippedKey struct {
	Path            string
	Value           string
	NearestMatches  []NearestMatch
}

// NearestMatch represents a potential catalog match that didn't meet the threshold
type NearestMatch struct {
	CatalogID       string
	CatalogName     string
	SimilarityScore float64
}

// ConversionResult contains the output and metadata from a conversion
type ConversionResult struct {
	OutputJSON         []byte
	SettingCount       int
	SkippedPayloads    []string
	SkippedKeys        []SkippedKey
	MatchedKeys        []MatchedKey
	WasSignedFile      bool
	OriginalData       []byte
	ProfileName        string
	ProfileDescription string
	UsedCustomConfig   bool
}

// OData type constants - Instance types (used in settingInstance)
const (
	ODataTypeGroupInstance           = "#microsoft.graph.deviceManagementConfigurationGroupSettingCollectionInstance"
	ODataTypeChoiceInstance          = "#microsoft.graph.deviceManagementConfigurationChoiceSettingInstance"
	ODataTypeSimpleInstance          = "#microsoft.graph.deviceManagementConfigurationSimpleSettingInstance"
	ODataTypeSimpleCollectionInstance = "#microsoft.graph.deviceManagementConfigurationSimpleSettingCollectionInstance"
	ODataTypeChoiceCollectionInstance = "#microsoft.graph.deviceManagementConfigurationChoiceSettingCollectionInstance"
)

// OData type constants - Value types (used in setting values)
const (
	ODataTypeStringValue  = "#microsoft.graph.deviceManagementConfigurationStringSettingValue"
	ODataTypeIntegerValue = "#microsoft.graph.deviceManagementConfigurationIntegerSettingValue"
	ODataTypeChoiceValue  = "#microsoft.graph.deviceManagementConfigurationChoiceSettingValue"
	ODataTypeSecretValue  = "#microsoft.graph.deviceManagementConfigurationSecretSettingValue"
)

// OData type constants - Definition types (used in catalog)
const (
	ODataTypeGroupCollectionDef  = "#microsoft.graph.deviceManagementConfigurationSettingGroupCollectionDefinition"
	ODataTypeSimpleCollectionDef = "#microsoft.graph.deviceManagementConfigurationSimpleSettingCollectionDefinition"
	ODataTypeChoiceCollectionDef = "#microsoft.graph.deviceManagementConfigurationChoiceSettingCollectionDefinition"
	ODataTypeChoiceDef           = "#microsoft.graph.deviceManagementConfigurationChoiceSettingDefinition"
	ODataTypeSimpleDef           = "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition"
)
