package utils

import (
	"fmt"
	"reflect"
	"time"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
}

// ValidationRule defines a validation rule
type ValidationRule interface {
	Validate(data interface{}) (bool, []ValidationError)
	Weight() float64
	GetName() string
}

// DataValidator validates data using multiple rules
type DataValidator struct {
	rules []ValidationRule
	score float64
}

// NewDataValidator creates a new data validator
func NewDataValidator() *DataValidator {
	return &DataValidator{
		rules: make([]ValidationRule, 0),
		score: 0.0,
	}
}

// AddRule adds a validation rule
func (dv *DataValidator) AddRule(rule ValidationRule) {
	dv.rules = append(dv.rules, rule)
}

// Validate validates data using all rules
func (dv *DataValidator) Validate(data interface{}) (bool, []ValidationError, float64) {
	var allErrors []ValidationError
	totalWeight := 0.0
	passedWeight := 0.0

	for _, rule := range dv.rules {
		valid, errors := rule.Validate(data)
		allErrors = append(allErrors, errors...)

		weight := rule.Weight()
		totalWeight += weight

		if valid {
			passedWeight += weight
		}
	}

	dv.score = 0.0
	if totalWeight > 0 {
		dv.score = passedWeight / totalWeight
	}

	return len(allErrors) == 0, allErrors, dv.score
}

// GetScore returns the current validation score
func (dv *DataValidator) GetScore() float64 {
	return dv.score
}

// RequiredFieldRule validates that a field is present and not empty
type RequiredFieldRule struct {
	fieldName string
	weight    float64
}

// NewRequiredFieldRule creates a new required field rule
func NewRequiredFieldRule(fieldName string, weight float64) *RequiredFieldRule {
	return &RequiredFieldRule{
		fieldName: fieldName,
		weight:    weight,
	}
}

// Validate checks if the field is present and not empty
func (r *RequiredFieldRule) Validate(data interface{}) (bool, []ValidationError) {
	value := getFieldValue(data, r.fieldName)

	if value == nil || (reflect.ValueOf(value).Kind() == reflect.String && value.(string) == "") {
		return false, []ValidationError{
			{
				Field:   r.fieldName,
				Message: "field is required",
				Value:   value,
			},
		}
	}

	return true, nil
}

// Weight returns the rule weight
func (r *RequiredFieldRule) Weight() float64 {
	return r.weight
}

// GetName returns the rule name
func (r *RequiredFieldRule) GetName() string {
	return fmt.Sprintf("required_field_%s", r.fieldName)
}

// RangeRule validates that a numeric field is within a range
type RangeRule struct {
	fieldName string
	min       float64
	max       float64
	weight    float64
}

// NewRangeRule creates a new range validation rule
func NewRangeRule(fieldName string, min, max, weight float64) *RangeRule {
	return &RangeRule{
		fieldName: fieldName,
		min:       min,
		max:       max,
		weight:    weight,
	}
}

// Validate checks if the field value is within the specified range
func (r *RangeRule) Validate(data interface{}) (bool, []ValidationError) {
	value := getFieldValue(data, r.fieldName)

	if value == nil {
		return true, nil // Skip if field is not present
	}

	var numValue float64
	switch v := value.(type) {
	case float64:
		numValue = v
	case float32:
		numValue = float64(v)
	case int:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	default:
		return false, []ValidationError{
			{
				Field:   r.fieldName,
				Message: "field is not numeric",
				Value:   value,
			},
		}
	}

	if numValue < r.min || numValue > r.max {
		return false, []ValidationError{
			{
				Field:   r.fieldName,
				Message: fmt.Sprintf("value must be between %f and %f", r.min, r.max),
				Value:   numValue,
			},
		}
	}

	return true, nil
}

// Weight returns the rule weight
func (r *RangeRule) Weight() float64 {
	return r.weight
}

// GetName returns the rule name
func (r *RangeRule) GetName() string {
	return fmt.Sprintf("range_%s", r.fieldName)
}

// TimeRangeRule validates that a time field is within a range
type TimeRangeRule struct {
	fieldName string
	start     time.Time
	end       time.Time
	weight    float64
}

// NewTimeRangeRule creates a new time range validation rule
func NewTimeRangeRule(fieldName string, start, end time.Time, weight float64) *TimeRangeRule {
	return &TimeRangeRule{
		fieldName: fieldName,
		start:     start,
		end:       end,
		weight:    weight,
	}
}

// Validate checks if the time field is within the specified range
func (r *TimeRangeRule) Validate(data interface{}) (bool, []ValidationError) {
	value := getFieldValue(data, r.fieldName)

	if value == nil {
		return true, nil // Skip if field is not present
	}

	var timeValue time.Time
	switch v := value.(type) {
	case time.Time:
		timeValue = v
	case string:
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return false, []ValidationError{
				{
					Field:   r.fieldName,
					Message: "invalid time format",
					Value:   value,
				},
			}
		}
		timeValue = parsed
	default:
		return false, []ValidationError{
			{
				Field:   r.fieldName,
				Message: "field is not a valid time",
				Value:   value,
			},
		}
	}

	if timeValue.Before(r.start) || timeValue.After(r.end) {
		return false, []ValidationError{
			{
				Field:   r.fieldName,
				Message: fmt.Sprintf("time must be between %s and %s", r.start, r.end),
				Value:   timeValue,
			},
		}
	}

	return true, nil
}

// Weight returns the rule weight
func (r *TimeRangeRule) Weight() float64 {
	return r.weight
}

// GetName returns the rule name
func (r *TimeRangeRule) GetName() string {
	return fmt.Sprintf("time_range_%s", r.fieldName)
}

// getFieldValue extracts a field value from a struct using reflection
func getFieldValue(data interface{}, fieldName string) interface{} {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return nil
	}

	return field.Interface()
}

// DeduplicateEarthquakes removes duplicate earthquakes based on ID and time
func DeduplicateEarthquakes(earthquakes []interface{}) []interface{} {
	seen := make(map[string]bool)
	unique := make([]interface{}, 0)

	for _, eq := range earthquakes {
		// Create a unique key based on ID and time
		key := createEarthquakeKey(eq)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, eq)
		}
	}

	return unique
}

// createEarthquakeKey creates a unique key for an earthquake
func createEarthquakeKey(eq interface{}) string {
	// This is a simplified implementation
	// In a real implementation, you would extract ID and time from the earthquake struct
	return fmt.Sprintf("%v", eq)
}

// ConvertValidationErrors converts validation errors to standard errors
func ConvertValidationErrors(validationErrors []ValidationError) []error {
	errors := make([]error, len(validationErrors))
	for i, ve := range validationErrors {
		errors[i] = &ve
	}
	return errors
}
