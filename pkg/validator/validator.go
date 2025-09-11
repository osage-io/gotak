package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Validator provides validation functionality
type Validator struct {
	errors []string
}

// New creates a new validator instance
func New() *Validator {
	return &Validator{
		errors: make([]string, 0),
	}
}

// ValidateStruct validates a struct based on struct tags
func (v *Validator) ValidateStruct(s interface{}) error {
	v.errors = make([]string, 0)
	
	val := reflect.ValueOf(s)
	typ := reflect.TypeOf(s)
	
	// Handle pointers
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}
	
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %v", val.Kind())
	}
	
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		
		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}
		
		v.validateField(field, fieldType)
	}
	
	if len(v.errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(v.errors, "; "))
	}
	
	return nil
}

func (v *Validator) validateField(field reflect.Value, fieldType reflect.StructField) {
	tag := fieldType.Tag.Get("validate")
	if tag == "" {
		return
	}
	
	fieldName := getFieldName(fieldType)
	rules := strings.Split(tag, ",")
	
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}
		
		v.applyRule(field, fieldName, rule)
	}
}

func (v *Validator) applyRule(field reflect.Value, fieldName, rule string) {
	parts := strings.SplitN(rule, "=", 2)
	ruleName := parts[0]
	var ruleValue string
	if len(parts) > 1 {
		ruleValue = parts[1]
	}
	
	switch ruleName {
	case "required":
		v.validateRequired(field, fieldName)
	case "min":
		v.validateMin(field, fieldName, ruleValue)
	case "max":
		v.validateMax(field, fieldName, ruleValue)
	case "len":
		v.validateLen(field, fieldName, ruleValue)
	case "email":
		v.validateEmail(field, fieldName)
	case "uuid":
		v.validateUUID(field, fieldName)
	case "oneof":
		v.validateOneOf(field, fieldName, ruleValue)
	default:
		v.addError(fieldName, fmt.Sprintf("unknown validation rule: %s", ruleName))
	}
}

func (v *Validator) validateRequired(field reflect.Value, fieldName string) {
	if isEmpty(field) {
		v.addError(fieldName, "is required")
	}
}

func (v *Validator) validateMin(field reflect.Value, fieldName, ruleValue string) {
	minVal, err := strconv.Atoi(ruleValue)
	if err != nil {
		v.addError(fieldName, fmt.Sprintf("invalid min rule value: %s", ruleValue))
		return
	}
	
	switch field.Kind() {
	case reflect.String:
		if field.Len() < minVal {
			v.addError(fieldName, fmt.Sprintf("must be at least %d characters long", minVal))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() < int64(minVal) {
			v.addError(fieldName, fmt.Sprintf("must be at least %d", minVal))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if field.Uint() < uint64(minVal) {
			v.addError(fieldName, fmt.Sprintf("must be at least %d", minVal))
		}
	case reflect.Slice, reflect.Array:
		if field.Len() < minVal {
			v.addError(fieldName, fmt.Sprintf("must have at least %d items", minVal))
		}
	}
}

func (v *Validator) validateMax(field reflect.Value, fieldName, ruleValue string) {
	maxVal, err := strconv.Atoi(ruleValue)
	if err != nil {
		v.addError(fieldName, fmt.Sprintf("invalid max rule value: %s", ruleValue))
		return
	}
	
	switch field.Kind() {
	case reflect.String:
		if field.Len() > maxVal {
			v.addError(fieldName, fmt.Sprintf("must be at most %d characters long", maxVal))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() > int64(maxVal) {
			v.addError(fieldName, fmt.Sprintf("must be at most %d", maxVal))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if field.Uint() > uint64(maxVal) {
			v.addError(fieldName, fmt.Sprintf("must be at most %d", maxVal))
		}
	case reflect.Slice, reflect.Array:
		if field.Len() > maxVal {
			v.addError(fieldName, fmt.Sprintf("must have at most %d items", maxVal))
		}
	}
}

func (v *Validator) validateLen(field reflect.Value, fieldName, ruleValue string) {
	lenVal, err := strconv.Atoi(ruleValue)
	if err != nil {
		v.addError(fieldName, fmt.Sprintf("invalid len rule value: %s", ruleValue))
		return
	}
	
	switch field.Kind() {
	case reflect.String:
		if field.Len() != lenVal {
			v.addError(fieldName, fmt.Sprintf("must be exactly %d characters long", lenVal))
		}
	case reflect.Slice, reflect.Array:
		if field.Len() != lenVal {
			v.addError(fieldName, fmt.Sprintf("must have exactly %d items", lenVal))
		}
	}
}

func (v *Validator) validateEmail(field reflect.Value, fieldName string) {
	if field.Kind() != reflect.String {
		v.addError(fieldName, "email validation can only be applied to strings")
		return
	}
	
	if isEmpty(field) {
		return // Skip if empty (use required rule for that)
	}
	
	email := field.String()
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		v.addError(fieldName, "must be a valid email address")
	}
}

func (v *Validator) validateUUID(field reflect.Value, fieldName string) {
	if field.Kind() != reflect.String {
		v.addError(fieldName, "uuid validation can only be applied to strings")
		return
	}
	
	if isEmpty(field) {
		return // Skip if empty (use required rule for that)
	}
	
	uuidStr := field.String()
	if _, err := uuid.Parse(uuidStr); err != nil {
		v.addError(fieldName, "must be a valid UUID")
	}
}

func (v *Validator) validateOneOf(field reflect.Value, fieldName, ruleValue string) {
	if field.Kind() != reflect.String {
		v.addError(fieldName, "oneof validation can only be applied to strings")
		return
	}
	
	if isEmpty(field) {
		return // Skip if empty (use required rule for that)
	}
	
	value := field.String()
	options := strings.Split(ruleValue, " ")
	
	for _, option := range options {
		if value == option {
			return
		}
	}
	
	v.addError(fieldName, fmt.Sprintf("must be one of: %s", strings.Join(options, ", ")))
}

func (v *Validator) addError(fieldName, message string) {
	v.errors = append(v.errors, fmt.Sprintf("%s %s", fieldName, message))
}

func isEmpty(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.String:
		return field.Len() == 0
	case reflect.Slice, reflect.Map, reflect.Array:
		return field.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return field.IsNil()
	case reflect.Invalid:
		return true
	default:
		return false
	}
}

func getFieldName(fieldType reflect.StructField) string {
	jsonTag := fieldType.Tag.Get("json")
	if jsonTag != "" && jsonTag != "-" {
		// Extract field name from json tag (before comma)
		if idx := strings.Index(jsonTag, ","); idx != -1 {
			return jsonTag[:idx]
		}
		return jsonTag
	}
	return fieldType.Name
}

// Custom validation functions

// ValidateTimeRange validates that end time is after start time
func ValidateTimeRange(start, end *time.Time, fieldName string) error {
	if start != nil && end != nil && end.Before(*start) {
		return fmt.Errorf("%s: end time must be after start time", fieldName)
	}
	return nil
}

// ValidatePriority validates priority is within valid range
func ValidatePriority(priority int, fieldName string) error {
	if priority < 1 || priority > 5 {
		return fmt.Errorf("%s: priority must be between 1 and 5", fieldName)
	}
	return nil
}

// ValidateEnum validates that value is in the allowed enum values
func ValidateEnum(value string, allowedValues []string, fieldName string) error {
	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}
	return fmt.Errorf("%s: must be one of %v", fieldName, allowedValues)
}
