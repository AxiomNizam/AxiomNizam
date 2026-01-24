package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TransformationRule defines how to transform data
type TransformationRule struct {
	Name             string                 `json:"name"`
	Description      string                 `json:"description,omitempty"`
	FieldMappings    map[string]string      `json:"field_mappings,omitempty"`    // source -> target field names
	TypeConversions  map[string]string      `json:"type_conversions,omitempty"`  // field -> target type
	FlattenConfig    *FlattenConfig         `json:"flatten_config,omitempty"`    // JSON flattening config
	Filters          []FilterRule           `json:"filters,omitempty"`           // Filter data before transformation
	AggregationRules map[string]string      `json:"aggregation_rules,omitempty"` // field -> aggregation type
	CustomTransforms map[string]interface{} `json:"custom_transforms,omitempty"` // Custom transformation functions
}

// FlattenConfig controls JSON flattening behavior
type FlattenConfig struct {
	Enabled       bool   `json:"enabled"`
	Separator     string `json:"separator"` // Default: "."
	Prefix        string `json:"prefix,omitempty"`
	MaxDepth      int    `json:"max_depth,omitempty"`       // 0 = unlimited
	SkipArrays    bool   `json:"skip_arrays,omitempty"`     // Skip flattening arrays
	SkipNullValue bool   `json:"skip_null_value,omitempty"` // Skip null/empty values
}

// FilterRule defines filtering conditions
type FilterRule struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, gte, lte, in, nin, contains, regex
	Value    interface{} `json:"value"`
}

// TransformedData holds the result of transformation
type TransformedData struct {
	Original    interface{}            `json:"original,omitempty"`
	Transformed interface{}            `json:"transformed"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Errors      []string               `json:"errors,omitempty"`
	FieldCount  int                    `json:"field_count"`
	Duration    string                 `json:"duration"`
}

// DataTransformer handles data transformations
type DataTransformer struct {
	rules map[string]*TransformationRule
}

// NewDataTransformer creates a new transformer instance
func NewDataTransformer() *DataTransformer {
	return &DataTransformer{
		rules: make(map[string]*TransformationRule),
	}
}

// RegisterRule registers a transformation rule
func (dt *DataTransformer) RegisterRule(rule *TransformationRule) error {
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}
	dt.rules[rule.Name] = rule
	return nil
}

// GetRule retrieves a registered rule
func (dt *DataTransformer) GetRule(name string) (*TransformationRule, error) {
	rule, exists := dt.rules[name]
	if !exists {
		return nil, fmt.Errorf("rule '%s' not found", name)
	}
	return rule, nil
}

// Transform applies a transformation rule to data
func (dt *DataTransformer) Transform(ruleName string, data interface{}) (*TransformedData, error) {
	startTime := time.Now()

	rule, err := dt.GetRule(ruleName)
	if err != nil {
		return nil, err
	}

	result := &TransformedData{
		Original: data,
		Metadata: make(map[string]interface{}),
		Errors:   []string{},
	}

	// Convert input to map if it's JSON
	var dataMap map[string]interface{}
	switch v := data.(type) {
	case map[string]interface{}:
		dataMap = v
	case string:
		err := json.Unmarshal([]byte(v), &dataMap)
		if err != nil {
			return result, fmt.Errorf("failed to parse JSON: %w", err)
		}
	default:
		// Convert struct to map
		jsonBytes, _ := json.Marshal(v)
		json.Unmarshal(jsonBytes, &dataMap)
	}

	if dataMap == nil {
		dataMap = make(map[string]interface{})
	}

	// Step 1: Apply filters
	if len(rule.Filters) > 0 {
		filtered, err := dt.applyFilters(dataMap, rule.Filters)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("filter error: %v", err))
		} else if !filtered {
			result.Transformed = nil
			result.Duration = time.Since(startTime).String()
			return result, nil
		}
	}

	// Step 2: Apply field mappings (rename fields)
	if rule.FieldMappings != nil && len(rule.FieldMappings) > 0 {
		dataMap = dt.applyFieldMappings(dataMap, rule.FieldMappings)
	}

	// Step 3: Apply type conversions
	if rule.TypeConversions != nil && len(rule.TypeConversions) > 0 {
		dataMap, errs := dt.applyTypeConversions(dataMap, rule.TypeConversions)
		if len(errs) > 0 {
			result.Errors = append(result.Errors, errs...)
		}
	}

	// Step 4: Apply JSON flattening
	if rule.FlattenConfig != nil && rule.FlattenConfig.Enabled {
		flattened, errs := dt.flattenJSON(dataMap, rule.FlattenConfig)
		if len(errs) > 0 {
			result.Errors = append(result.Errors, errs...)
		}
		dataMap = flattened
	}

	// Step 5: Apply aggregation rules
	if rule.AggregationRules != nil && len(rule.AggregationRules) > 0 {
		dataMap, errs := dt.applyAggregations(dataMap, rule.AggregationRules)
		if len(errs) > 0 {
			result.Errors = append(result.Errors, errs...)
		}
	}

	result.Transformed = dataMap
	result.FieldCount = len(flattenKeysCount(dataMap))
	result.Duration = time.Since(startTime).String()

	return result, nil
}

// TransformBatch transforms multiple data items
func (dt *DataTransformer) TransformBatch(ruleName string, dataList []interface{}) ([]*TransformedData, error) {
	results := make([]*TransformedData, 0, len(dataList))

	for _, data := range dataList {
		result, err := dt.Transform(ruleName, data)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
		}
		results = append(results, result)
	}

	return results, nil
}

// applyFilters checks if data matches all filter conditions
func (dt *DataTransformer) applyFilters(data map[string]interface{}, filters []FilterRule) (bool, error) {
	for _, filter := range filters {
		match, err := dt.evaluateFilter(data, filter)
		if err != nil {
			return false, err
		}
		if !match {
			return false, nil
		}
	}
	return true, nil
}

// evaluateFilter checks a single filter condition
func (dt *DataTransformer) evaluateFilter(data map[string]interface{}, filter FilterRule) (bool, error) {
	value, exists := data[filter.Field]
	if !exists {
		return filter.Operator == "ne", nil
	}

	switch filter.Operator {
	case "eq":
		return value == filter.Value, nil
	case "ne":
		return value != filter.Value, nil
	case "gt":
		return compareValues(value, filter.Value) > 0, nil
	case "lt":
		return compareValues(value, filter.Value) < 0, nil
	case "gte":
		return compareValues(value, filter.Value) >= 0, nil
	case "lte":
		return compareValues(value, filter.Value) <= 0, nil
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", value), fmt.Sprintf("%v", filter.Value)), nil
	case "in":
		return inArray(value, filter.Value), nil
	case "nin":
		return !inArray(value, filter.Value), nil
	default:
		return false, fmt.Errorf("unknown operator: %s", filter.Operator)
	}
}

// applyFieldMappings renames fields according to mapping rules
func (dt *DataTransformer) applyFieldMappings(data map[string]interface{}, mappings map[string]string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		// Check if this field has a mapping
		if newKey, exists := mappings[key]; exists {
			result[newKey] = value
		} else {
			result[key] = value
		}
	}

	return result
}

// applyTypeConversions converts field types
func (dt *DataTransformer) applyTypeConversions(data map[string]interface{}, conversions map[string]string) (map[string]interface{}, []string) {
	var errors []string
	result := make(map[string]interface{})

	for key, value := range data {
		if targetType, exists := conversions[key]; exists {
			converted, err := dt.convertType(value, targetType)
			if err != nil {
				errors = append(errors, fmt.Sprintf("field '%s': %v", key, err))
				result[key] = value // Keep original on error
			} else {
				result[key] = converted
			}
		} else {
			result[key] = value
		}
	}

	return result, errors
}

// convertType converts a value to a target type
func (dt *DataTransformer) convertType(value interface{}, targetType string) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch strings.ToLower(targetType) {
	case "string":
		return fmt.Sprintf("%v", value), nil
	case "int", "integer":
		switch v := value.(type) {
		case float64:
			return int64(v), nil
		case string:
			num, err := strconv.ParseInt(v, 10, 64)
			return num, err
		case int:
			return int64(v), nil
		default:
			return nil, fmt.Errorf("cannot convert %T to int", value)
		}
	case "float", "double":
		switch v := value.(type) {
		case float64:
			return v, nil
		case string:
			return strconv.ParseFloat(v, 64)
		case int:
			return float64(v), nil
		default:
			return nil, fmt.Errorf("cannot convert %T to float", value)
		}
	case "bool", "boolean":
		switch v := value.(type) {
		case bool:
			return v, nil
		case string:
			return strings.ToLower(v) == "true", nil
		default:
			return nil, fmt.Errorf("cannot convert %T to bool", value)
		}
	case "timestamp":
		switch v := value.(type) {
		case string:
			t, err := time.Parse(time.RFC3339, v)
			return t.Unix(), err
		case float64:
			return int64(v), nil
		default:
			return nil, fmt.Errorf("cannot convert %T to timestamp", value)
		}
	case "date":
		switch v := value.(type) {
		case string:
			t, err := time.Parse("2006-01-02", v)
			return t.Format("2006-01-02"), err
		default:
			return nil, fmt.Errorf("cannot convert %T to date", value)
		}
	default:
		return value, nil
	}
}

// flattenJSON flattens nested JSON structure
func (dt *DataTransformer) flattenJSON(data interface{}, config *FlattenConfig) (map[string]interface{}, []string) {
	var errors []string
	separator := "."
	if config.Separator != "" {
		separator = config.Separator
	}

	prefix := ""
	if config.Prefix != "" {
		prefix = config.Prefix + separator
	}

	result := make(map[string]interface{})
	dt.flattenRecursive(data, prefix, result, separator, config, 0)

	return result, errors
}

// flattenRecursive recursively flattens nested structures
func (dt *DataTransformer) flattenRecursive(data interface{}, prefix string, result map[string]interface{}, separator string, config *FlattenConfig, depth int) {
	if config.MaxDepth > 0 && depth >= config.MaxDepth {
		if data != nil && !config.SkipNullValue {
			result[strings.TrimSuffix(prefix, separator)] = data
		}
		return
	}

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			newKey := prefix + key
			if value == nil && config.SkipNullValue {
				continue
			}

			switch value.(type) {
			case map[string]interface{}:
				dt.flattenRecursive(value, newKey+separator, result, separator, config, depth+1)
			case []interface{}:
				if config.SkipArrays {
					result[newKey] = value
				} else {
					dt.flattenRecursive(value, newKey+separator, result, separator, config, depth+1)
				}
			default:
				result[newKey] = value
			}
		}
	case []interface{}:
		for idx, item := range v {
			newKey := prefix + strconv.Itoa(idx)
			if item == nil && config.SkipNullValue {
				continue
			}

			switch item.(type) {
			case map[string]interface{}:
				dt.flattenRecursive(item, newKey+separator, result, separator, config, depth+1)
			case []interface{}:
				if config.SkipArrays {
					result[newKey] = item
				} else {
					dt.flattenRecursive(item, newKey+separator, result, separator, config, depth+1)
				}
			default:
				result[newKey] = item
			}
		}
	default:
		key := strings.TrimSuffix(prefix, separator)
		if v != nil && !config.SkipNullValue {
			result[key] = data
		} else if v != nil {
			result[key] = data
		}
	}
}

// applyAggregations applies aggregation functions to data
func (dt *DataTransformer) applyAggregations(data map[string]interface{}, rules map[string]string) (map[string]interface{}, []string) {
	var errors []string
	result := make(map[string]interface{})

	for key, value := range data {
		if aggType, exists := rules[key]; exists {
			agg, err := dt.applyAggregation(value, aggType)
			if err != nil {
				errors = append(errors, fmt.Sprintf("aggregation on '%s': %v", key, err))
				result[key] = value
			} else {
				result[key] = agg
			}
		} else {
			result[key] = value
		}
	}

	return result, errors
}

// applyAggregation applies a single aggregation function
func (dt *DataTransformer) applyAggregation(value interface{}, aggType string) (interface{}, error) {
	arrVal, ok := value.([]interface{})
	if !ok {
		return value, nil
	}

	switch strings.ToLower(aggType) {
	case "sum":
		total := 0.0
		for _, item := range arrVal {
			if num, err := toNumber(item); err == nil {
				total += num
			}
		}
		return total, nil
	case "avg", "average":
		if len(arrVal) == 0 {
			return 0, nil
		}
		total := 0.0
		for _, item := range arrVal {
			if num, err := toNumber(item); err == nil {
				total += num
			}
		}
		return total / float64(len(arrVal)), nil
	case "count":
		return len(arrVal), nil
	case "min":
		if len(arrVal) == 0 {
			return nil, nil
		}
		minVal, _ := toNumber(arrVal[0])
		for _, item := range arrVal[1:] {
			if num, err := toNumber(item); err == nil && num < minVal {
				minVal = num
			}
		}
		return minVal, nil
	case "max":
		if len(arrVal) == 0 {
			return nil, nil
		}
		maxVal, _ := toNumber(arrVal[0])
		for _, item := range arrVal[1:] {
			if num, err := toNumber(item); err == nil && num > maxVal {
				maxVal = num
			}
		}
		return maxVal, nil
	case "join":
		var strs []string
		for _, item := range arrVal {
			strs = append(strs, fmt.Sprintf("%v", item))
		}
		return strings.Join(strs, ","), nil
	default:
		return value, nil
	}
}

// Helper functions

func compareValues(a, b interface{}) int {
	aNum, _ := toNumber(a)
	bNum, _ := toNumber(b)
	if aNum > bNum {
		return 1
	} else if aNum < bNum {
		return -1
	}
	return 0
}

func toNumber(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to number", v)
	}
}

func inArray(value interface{}, arr interface{}) bool {
	switch v := arr.(type) {
	case []interface{}:
		for _, item := range v {
			if item == value {
				return true
			}
		}
	}
	return false
}

func flattenKeysCount(data map[string]interface{}) map[string]bool {
	keys := make(map[string]bool)
	for key := range data {
		keys[key] = true
	}
	return keys
}

// ListRules returns all registered rule names
func (dt *DataTransformer) ListRules() []string {
	names := make([]string, 0, len(dt.rules))
	for name := range dt.rules {
		names = append(names, name)
	}
	return names
}

// DeleteRule removes a registered rule
func (dt *DataTransformer) DeleteRule(name string) error {
	if _, exists := dt.rules[name]; !exists {
		return fmt.Errorf("rule '%s' not found", name)
	}
	delete(dt.rules, name)
	return nil
}

// ExportRules exports all rules as JSON
func (dt *DataTransformer) ExportRules() ([]byte, error) {
	return json.MarshalIndent(dt.rules, "", "  ")
}

// ImportRules imports rules from JSON
func (dt *DataTransformer) ImportRules(jsonData []byte) error {
	var rules map[string]*TransformationRule
	if err := json.Unmarshal(jsonData, &rules); err != nil {
		return err
	}

	for _, rule := range rules {
		if err := dt.RegisterRule(rule); err != nil {
			return err
		}
	}

	return nil
}
