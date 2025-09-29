package rbac

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// jsonLogicEvaluator implements the AttributeEvaluator interface using JSON Logic
type jsonLogicEvaluator struct{}

// NewAttributeEvaluator creates a new JSON Logic attribute evaluator
func NewAttributeEvaluator() AttributeEvaluator {
	return &jsonLogicEvaluator{}
}

// Evaluate evaluates a JSON Logic expression against provided attributes
func (e *jsonLogicEvaluator) Evaluate(expression string, attributes map[string]interface{}) (bool, error) {
	// Parse the JSON Logic expression
	var rule map[string]interface{}
	if err := json.Unmarshal([]byte(expression), &rule); err != nil {
		return false, fmt.Errorf("invalid JSON expression: %v", err)
	}

	// Evaluate the rule
	result, err := e.evaluateRule(rule, attributes)
	if err != nil {
		return false, err
	}

	// Convert result to boolean
	return e.toBool(result), nil
}

// ValidateExpression validates that a JSON Logic expression is valid
func (e *jsonLogicEvaluator) ValidateExpression(expression string) error {
	var rule map[string]interface{}
	if err := json.Unmarshal([]byte(expression), &rule); err != nil {
		return fmt.Errorf("invalid JSON expression: %v", err)
	}

	// Validate the rule structure
	return e.validateRule(rule)
}

// ExtractRequiredAttributes extracts the required attributes from an expression
func (e *jsonLogicEvaluator) ExtractRequiredAttributes(expression string) ([]string, error) {
	var rule map[string]interface{}
	if err := json.Unmarshal([]byte(expression), &rule); err != nil {
		return nil, fmt.Errorf("invalid JSON expression: %v", err)
	}

	attributes := make(map[string]bool)
	e.extractAttributes(rule, attributes)

	result := make([]string, 0, len(attributes))
	for attr := range attributes {
		result = append(result, attr)
	}

	return result, nil
}

// evaluateRule evaluates a JSON Logic rule
func (e *jsonLogicEvaluator) evaluateRule(rule map[string]interface{}, data map[string]interface{}) (interface{}, error) {
	// Handle empty rule
	if len(rule) == 0 {
		return nil, nil
	}

	// Get the operator and arguments
	for op, args := range rule {
		return e.evaluateOperation(op, args, data)
	}

	return nil, fmt.Errorf("invalid rule structure")
}

// evaluateOperation evaluates a specific JSON Logic operation
func (e *jsonLogicEvaluator) evaluateOperation(operator string, args interface{}, data map[string]interface{}) (interface{}, error) {
	switch operator {
	case "var":
		return e.evaluateVar(args, data)
	case "==", "===":
		return e.evaluateEquals(args, data, true)
	case "!=", "!==":
		return e.evaluateEquals(args, data, false)
	case ">":
		return e.evaluateComparison(args, data, ">")
	case ">=":
		return e.evaluateComparison(args, data, ">=")
	case "<":
		return e.evaluateComparison(args, data, "<")
	case "<=":
		return e.evaluateComparison(args, data, "<=")
	case "!":
		return e.evaluateNot(args, data)
	case "and":
		return e.evaluateAnd(args, data)
	case "or":
		return e.evaluateOr(args, data)
	case "if":
		return e.evaluateIf(args, data)
	case "in":
		return e.evaluateIn(args, data)
	case "cat":
		return e.evaluateCat(args, data)
	case "+":
		return e.evaluateAdd(args, data)
	case "-":
		return e.evaluateSubtract(args, data)
	case "*":
		return e.evaluateMultiply(args, data)
	case "/":
		return e.evaluateDivide(args, data)
	case "%":
		return e.evaluateModulo(args, data)
	case "min":
		return e.evaluateMin(args, data)
	case "max":
		return e.evaluateMax(args, data)
	case "map":
		return e.evaluateMap(args, data)
	case "filter":
		return e.evaluateFilter(args, data)
	case "reduce":
		return e.evaluateReduce(args, data)
	case "all":
		return e.evaluateAll(args, data)
	case "none":
		return e.evaluateNone(args, data)
	case "some":
		return e.evaluateSome(args, data)
	case "substr":
		return e.evaluateSubstr(args, data)
	case "length":
		return e.evaluateLength(args, data)
	case "regex":
		return e.evaluateRegex(args, data)
	case "time":
		return e.evaluateTime(args, data)
	default:
		return nil, fmt.Errorf("unknown operator: %s", operator)
	}
}

// evaluateVar gets a variable from data
func (e *jsonLogicEvaluator) evaluateVar(args interface{}, data map[string]interface{}) (interface{}, error) {
	argsList, err := e.toSlice(args)
	if err != nil {
		return nil, err
	}

	if len(argsList) == 0 {
		return data, nil
	}

	path := e.toString(argsList[0])
	defaultVal := interface{}(nil)
	if len(argsList) > 1 {
		defaultVal = argsList[1]
	}

	// Handle nested paths with dot notation
	if strings.Contains(path, ".") {
		return e.getNestedValue(data, strings.Split(path, "."), defaultVal), nil
	}

	if val, exists := data[path]; exists {
		return val, nil
	}

	return defaultVal, nil
}

// evaluateEquals evaluates equality operations
func (e *jsonLogicEvaluator) evaluateEquals(args interface{}, data map[string]interface{}, expected bool) (interface{}, error) {
	argsList, err := e.toSlice(args)
	if err != nil {
		return nil, err
	}

	if len(argsList) < 2 {
		return nil, fmt.Errorf("equals operation requires at least 2 arguments")
	}

	val1, err := e.evaluateArg(argsList[0], data)
	if err != nil {
		return nil, err
	}

	val2, err := e.evaluateArg(argsList[1], data)
	if err != nil {
		return nil, err
	}

	equal := e.deepEqual(val1, val2)
	if expected {
		return equal, nil
	}
	return !equal, nil
}

// evaluateComparison evaluates comparison operations
func (e *jsonLogicEvaluator) evaluateComparison(args interface{}, data map[string]interface{}, operator string) (interface{}, error) {
	argsList, err := e.toSlice(args)
	if err != nil {
		return nil, err
	}

	if len(argsList) < 2 {
		return nil, fmt.Errorf("comparison operation requires at least 2 arguments")
	}

	val1, err := e.evaluateArg(argsList[0], data)
	if err != nil {
		return nil, err
	}

	val2, err := e.evaluateArg(argsList[1], data)
	if err != nil {
		return nil, err
	}

	num1, ok1 := e.toFloat64(val1)
	num2, ok2 := e.toFloat64(val2)

	if !ok1 || !ok2 {
		// String comparison as fallback
		str1 := e.toString(val1)
		str2 := e.toString(val2)
		
		switch operator {
		case ">":
			return str1 > str2, nil
		case ">=":
			return str1 >= str2, nil
		case "<":
			return str1 < str2, nil
		case "<=":
			return str1 <= str2, nil
		}
	}

	switch operator {
	case ">":
		return num1 > num2, nil
	case ">=":
		return num1 >= num2, nil
	case "<":
		return num1 < num2, nil
	case "<=":
		return num1 <= num2, nil
	}

	return false, nil
}

// evaluateNot evaluates logical NOT
func (e *jsonLogicEvaluator) evaluateNot(args interface{}, data map[string]interface{}) (interface{}, error) {
	val, err := e.evaluateArg(args, data)
	if err != nil {
		return nil, err
	}

	return !e.toBool(val), nil
}

// evaluateAnd evaluates logical AND
func (e *jsonLogicEvaluator) evaluateAnd(args interface{}, data map[string]interface{}) (interface{}, error) {
	argsList, err := e.toSlice(args)
	if err != nil {
		return nil, err
	}

	for _, arg := range argsList {
		val, err := e.evaluateArg(arg, data)
		if err != nil {
			return nil, err
		}

		if !e.toBool(val) {
			return false, nil
		}
	}

	return true, nil
}

// evaluateOr evaluates logical OR
func (e *jsonLogicEvaluator) evaluateOr(args interface{}, data map[string]interface{}) (interface{}, error) {
	argsList, err := e.toSlice(args)
	if err != nil {
		return nil, err
	}

	for _, arg := range argsList {
		val, err := e.evaluateArg(arg, data)
		if err != nil {
			return nil, err
		}

		if e.toBool(val) {
			return true, nil
		}
	}

	return false, nil
}

// evaluateIf evaluates conditional logic
func (e *jsonLogicEvaluator) evaluateIf(args interface{}, data map[string]interface{}) (interface{}, error) {
	argsList, err := e.toSlice(args)
	if err != nil {
		return nil, err
	}

	if len(argsList) < 2 {
		return nil, fmt.Errorf("if operation requires at least 2 arguments")
	}

	condition, err := e.evaluateArg(argsList[0], data)
	if err != nil {
		return nil, err
	}

	if e.toBool(condition) {
		return e.evaluateArg(argsList[1], data)
	}

	if len(argsList) > 2 {
		return e.evaluateArg(argsList[2], data)
	}

	return nil, nil
}

// evaluateIn checks if a value is in an array
func (e *jsonLogicEvaluator) evaluateIn(args interface{}, data map[string]interface{}) (interface{}, error) {
	argsList, err := e.toSlice(args)
	if err != nil {
		return nil, err
	}

	if len(argsList) < 2 {
		return nil, fmt.Errorf("in operation requires 2 arguments")
	}

	needle, err := e.evaluateArg(argsList[0], data)
	if err != nil {
		return nil, err
	}

	haystack, err := e.evaluateArg(argsList[1], data)
	if err != nil {
		return nil, err
	}

	haystackSlice, ok := haystack.([]interface{})
	if !ok {
		// Check if it's a string
		if haystackStr, ok := haystack.(string); ok {
			needleStr := e.toString(needle)
			return strings.Contains(haystackStr, needleStr), nil
		}
		return false, nil
	}

	for _, item := range haystackSlice {
		if e.deepEqual(needle, item) {
			return true, nil
		}
	}

	return false, nil
}

// evaluateCat concatenates strings
func (e *jsonLogicEvaluator) evaluateCat(args interface{}, data map[string]interface{}) (interface{}, error) {
	argsList, err := e.toSlice(args)
	if err != nil {
		return nil, err
	}

	var result strings.Builder
	for _, arg := range argsList {
		val, err := e.evaluateArg(arg, data)
		if err != nil {
			return nil, err
		}
		result.WriteString(e.toString(val))
	}

	return result.String(), nil
}

// evaluateAdd adds numbers
func (e *jsonLogicEvaluator) evaluateAdd(args interface{}, data map[string]interface{}) (interface{}, error) {
	argsList, err := e.toSlice(args)
	if err != nil {
		return nil, err
	}

	var sum float64
	for _, arg := range argsList {
		val, err := e.evaluateArg(arg, data)
		if err != nil {
			return nil, err
		}

		num, ok := e.toFloat64(val)
		if !ok {
			return nil, fmt.Errorf("cannot add non-numeric value: %v", val)
		}
		sum += num
	}

	// Return int if the result is a whole number
	if sum == float64(int64(sum)) {
		return int64(sum), nil
	}
	return sum, nil
}

// evaluateSubtract subtracts numbers
func (e *jsonLogicEvaluator) evaluateSubtract(args interface{}, data map[string]interface{}) (interface{}, error) {
	argsList, err := e.toSlice(args)
	if err != nil {
		return nil, err
	}

	if len(argsList) == 0 {
		return 0, nil
	}

	first, err := e.evaluateArg(argsList[0], data)
	if err != nil {
		return nil, err
	}

	result, ok := e.toFloat64(first)
	if !ok {
		return nil, fmt.Errorf("cannot subtract from non-numeric value: %v", first)
	}

	if len(argsList) == 1 {
		return -result, nil
	}

	for _, arg := range argsList[1:] {
		val, err := e.evaluateArg(arg, data)
		if err != nil {
			return nil, err
		}

		num, ok := e.toFloat64(val)
		if !ok {
			return nil, fmt.Errorf("cannot subtract non-numeric value: %v", val)
		}
		result -= num
	}

	if result == float64(int64(result)) {
		return int64(result), nil
	}
	return result, nil
}

// evaluateMultiply multiplies numbers
func (e *jsonLogicEvaluator) evaluateMultiply(args interface{}, data map[string]interface{}) (interface{}, error) {
	argsList, err := e.toSlice(args)
	if err != nil {
		return nil, err
	}

	result := 1.0
	for _, arg := range argsList {
		val, err := e.evaluateArg(arg, data)
		if err != nil {
			return nil, err
		}

		num, ok := e.toFloat64(val)
		if !ok {
			return nil, fmt.Errorf("cannot multiply non-numeric value: %v", val)
		}
		result *= num
	}

	if result == float64(int64(result)) {
		return int64(result), nil
	}
	return result, nil
}

// evaluateDivide divides numbers
func (e *jsonLogicEvaluator) evaluateDivide(args interface{}, data map[string]interface{}) (interface{}, error) {
	argsList, err := e.toSlice(args)
	if err != nil {
		return nil, err
	}

	if len(argsList) != 2 {
		return nil, fmt.Errorf("divide operation requires exactly 2 arguments")
	}

	numerator, err := e.evaluateArg(argsList[0], data)
	if err != nil {
		return nil, err
	}

	denominator, err := e.evaluateArg(argsList[1], data)
	if err != nil {
		return nil, err
	}

	num1, ok1 := e.toFloat64(numerator)
	num2, ok2 := e.toFloat64(denominator)

	if !ok1 || !ok2 {
		return nil, fmt.Errorf("cannot divide non-numeric values")
	}

	if num2 == 0 {
		return nil, fmt.Errorf("division by zero")
	}

	result := num1 / num2
	if result == float64(int64(result)) {
		return int64(result), nil
	}
	return result, nil
}

// evaluateModulo calculates modulo
func (e *jsonLogicEvaluator) evaluateModulo(args interface{}, data map[string]interface{}) (interface{}, error) {
	argsList, err := e.toSlice(args)
	if err != nil {
		return nil, err
	}

	if len(argsList) != 2 {
		return nil, fmt.Errorf("modulo operation requires exactly 2 arguments")
	}

	val1, err := e.evaluateArg(argsList[0], data)
	if err != nil {
		return nil, err
	}

	val2, err := e.evaluateArg(argsList[1], data)
	if err != nil {
		return nil, err
	}

	num1, ok1 := e.toFloat64(val1)
	num2, ok2 := e.toFloat64(val2)

	if !ok1 || !ok2 {
		return nil, fmt.Errorf("cannot calculate modulo of non-numeric values")
	}

	if num2 == 0 {
		return nil, fmt.Errorf("modulo by zero")
	}

	// Convert to integers for modulo operation
	int1 := int64(num1)
	int2 := int64(num2)
	return int1 % int2, nil
}

// evaluateMin finds minimum value
func (e *jsonLogicEvaluator) evaluateMin(args interface{}, data map[string]interface{}) (interface{}, error) {
	argsList, err := e.toSlice(args)
	if err != nil {
		return nil, err
	}

	if len(argsList) == 0 {
		return nil, nil
	}

	minVal, err := e.evaluateArg(argsList[0], data)
	if err != nil {
		return nil, err
	}

	minNum, ok := e.toFloat64(minVal)
	if !ok {
		return nil, fmt.Errorf("cannot find min of non-numeric value: %v", minVal)
	}

	for _, arg := range argsList[1:] {
		val, err := e.evaluateArg(arg, data)
		if err != nil {
			return nil, err
		}

		num, ok := e.toFloat64(val)
		if !ok {
			return nil, fmt.Errorf("cannot find min of non-numeric value: %v", val)
		}

		if num < minNum {
			minNum = num
			minVal = val
		}
	}

	return minVal, nil
}

// evaluateMax finds maximum value
func (e *jsonLogicEvaluator) evaluateMax(args interface{}, data map[string]interface{}) (interface{}, error) {
	argsList, err := e.toSlice(args)
	if err != nil {
		return nil, err
	}

	if len(argsList) == 0 {
		return nil, nil
	}

	maxVal, err := e.evaluateArg(argsList[0], data)
	if err != nil {
		return nil, err
	}

	maxNum, ok := e.toFloat64(maxVal)
	if !ok {
		return nil, fmt.Errorf("cannot find max of non-numeric value: %v", maxVal)
	}

	for _, arg := range argsList[1:] {
		val, err := e.evaluateArg(arg, data)
		if err != nil {
			return nil, err
		}

		num, ok := e.toFloat64(val)
		if !ok {
			return nil, fmt.Errorf("cannot find max of non-numeric value: %v", val)
		}

		if num > maxNum {
			maxNum = num
			maxVal = val
		}
	}

	return maxVal, nil
}

// Placeholder implementations for complex operations
func (e *jsonLogicEvaluator) evaluateMap(args interface{}, data map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("map operation not yet implemented")
}

func (e *jsonLogicEvaluator) evaluateFilter(args interface{}, data map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("filter operation not yet implemented")
}

func (e *jsonLogicEvaluator) evaluateReduce(args interface{}, data map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("reduce operation not yet implemented")
}

func (e *jsonLogicEvaluator) evaluateAll(args interface{}, data map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("all operation not yet implemented")
}

func (e *jsonLogicEvaluator) evaluateNone(args interface{}, data map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("none operation not yet implemented")
}

func (e *jsonLogicEvaluator) evaluateSome(args interface{}, data map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("some operation not yet implemented")
}

func (e *jsonLogicEvaluator) evaluateSubstr(args interface{}, data map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("substr operation not yet implemented")
}

func (e *jsonLogicEvaluator) evaluateLength(args interface{}, data map[string]interface{}) (interface{}, error) {
	val, err := e.evaluateArg(args, data)
	if err != nil {
		return nil, err
	}

	switch v := val.(type) {
	case string:
		return len(v), nil
	case []interface{}:
		return len(v), nil
	case map[string]interface{}:
		return len(v), nil
	default:
		return 0, nil
	}
}

func (e *jsonLogicEvaluator) evaluateRegex(args interface{}, data map[string]interface{}) (interface{}, error) {
	argsList, err := e.toSlice(args)
	if err != nil {
		return nil, err
	}

	if len(argsList) != 2 {
		return nil, fmt.Errorf("regex operation requires exactly 2 arguments")
	}

	text, err := e.evaluateArg(argsList[0], data)
	if err != nil {
		return nil, err
	}

	pattern, err := e.evaluateArg(argsList[1], data)
	if err != nil {
		return nil, err
	}

	textStr := e.toString(text)
	patternStr := e.toString(pattern)

	matched, err := regexp.MatchString(patternStr, textStr)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %v", err)
	}

	return matched, nil
}

func (e *jsonLogicEvaluator) evaluateTime(args interface{}, data map[string]interface{}) (interface{}, error) {
	return time.Now().Unix(), nil
}

// Helper methods

func (e *jsonLogicEvaluator) evaluateArg(arg interface{}, data map[string]interface{}) (interface{}, error) {
	// If it's a map, it's another rule to evaluate
	if rule, ok := arg.(map[string]interface{}); ok {
		return e.evaluateRule(rule, data)
	}
	
	// Otherwise, it's a literal value
	return arg, nil
}

func (e *jsonLogicEvaluator) validateRule(rule map[string]interface{}) error {
	for op := range rule {
		switch op {
		case "var", "==", "===", "!=", "!==", ">", ">=", "<", "<=", "!",
			 "and", "or", "if", "in", "cat", "+", "-", "*", "/", "%",
			 "min", "max", "length", "regex", "time":
			// Valid operators
		default:
			return fmt.Errorf("unknown operator: %s", op)
		}
	}
	return nil
}

func (e *jsonLogicEvaluator) extractAttributes(rule interface{}, attributes map[string]bool) {
	switch r := rule.(type) {
	case map[string]interface{}:
		for op, args := range r {
			if op == "var" {
				if argsList, err := e.toSlice(args); err == nil && len(argsList) > 0 {
					if varName, ok := argsList[0].(string); ok {
						attributes[varName] = true
					}
				}
			} else {
				e.extractAttributes(args, attributes)
			}
		}
	case []interface{}:
		for _, item := range r {
			e.extractAttributes(item, attributes)
		}
	}
}

func (e *jsonLogicEvaluator) getNestedValue(data map[string]interface{}, path []string, defaultVal interface{}) interface{} {
	current := data
	for i, key := range path {
		if i == len(path)-1 {
			if val, exists := current[key]; exists {
				return val
			}
			return defaultVal
		}

		if next, exists := current[key]; exists {
			if nextMap, ok := next.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return defaultVal
			}
		} else {
			return defaultVal
		}
	}

	return defaultVal
}

func (e *jsonLogicEvaluator) toSlice(v interface{}) ([]interface{}, error) {
	switch val := v.(type) {
	case []interface{}:
		return val, nil
	case nil:
		return []interface{}{}, nil
	default:
		return []interface{}{val}, nil
	}
}

func (e *jsonLogicEvaluator) toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}

func (e *jsonLogicEvaluator) toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float64:
		return val, true
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
		return 0, false
	default:
		return 0, false
	}
}

func (e *jsonLogicEvaluator) toBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val != 0
	case int64:
		return val != 0
	case float64:
		return val != 0
	case string:
		return val != ""
	case nil:
		return false
	default:
		return true
	}
}

func (e *jsonLogicEvaluator) deepEqual(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}
