package less_go

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func floatsAlmostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestUnitConversionsMatchJS(t *testing.T) {
	jsPath := filepath.Join("..", "..", "less", "data", "unit-conversions.js")
	jsContentBytes, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read unit-conversions.js from path %s: %v", jsPath, err)
	}
	jsContent := string(jsContentBytes)

	re := regexp.MustCompile(`export default\s*(\{[\s\S]*\});`)
	matches := re.FindStringSubmatch(jsContent)
	if len(matches) < 2 {
		t.Fatalf("Could not extract object from unit-conversions.js content:\n%s", jsContent)
	}
	jsObjectString := matches[1]

	jsObjectString = strings.ReplaceAll(jsObjectString, "Math.PI", fmt.Sprintf("%.15f", math.Pi))
	reMulSimple := regexp.MustCompile(`(\d+(\.\d+)?)\s*\*\s*(\d+(\.\d+)?)`)
	reDivSimple := regexp.MustCompile(`(\d+(\.\d+)?)\s*/\s*(\d+(\.\d+)?)`)
	reParenNum := regexp.MustCompile(`\(\s*(\d+(\.\d+)?)\s*\)`)

	for {
		mulIdx := reMulSimple.FindStringIndex(jsObjectString)
		divIdx := reDivSimple.FindStringIndex(jsObjectString)
		parenIdx := reParenNum.FindStringIndex(jsObjectString)

		firstIdx := -1
		operationType := ""

		if parenIdx != nil {
			firstIdx = parenIdx[0]
			operationType = "paren"
		}

		if mulIdx != nil && (firstIdx == -1 || mulIdx[0] < firstIdx) {
			firstIdx = mulIdx[0]
			operationType = "mul"
		}

		if divIdx != nil && (firstIdx == -1 || divIdx[0] < firstIdx) {
			firstIdx = divIdx[0]
			operationType = "div"
		}

		if operationType == "" {
			break
		}

		switch operationType {
		case "paren":
			match := reParenNum.FindStringSubmatch(jsObjectString)
			if len(match) > 1 {
				jsObjectString = strings.Replace(jsObjectString, match[0], match[1], 1)
			}
		case "mul":
			match := reMulSimple.FindStringSubmatch(jsObjectString)
			if len(match) > 3 {
				num1, err1 := strconv.ParseFloat(match[1], 64)
				num2, err2 := strconv.ParseFloat(match[3], 64)
				if err1 == nil && err2 == nil {
					result := fmt.Sprintf("%.15f", num1*num2)
					jsObjectString = strings.Replace(jsObjectString, match[0], result, 1)
				}
			}
		case "div":
			match := reDivSimple.FindStringSubmatch(jsObjectString)
			if len(match) > 3 {
				num, err1 := strconv.ParseFloat(match[1], 64)
				den, err2 := strconv.ParseFloat(match[3], 64)
				if err1 == nil && err2 == nil && den != 0 {
					result := fmt.Sprintf("%.15f", num/den)
					jsObjectString = strings.Replace(jsObjectString, match[0], result, 1)
				} else if den == 0 {
					jsObjectString = strings.Replace(jsObjectString, match[0], "0", 1)
				}
			}
		}
	}

	jsObjectString = regexp.MustCompile(`([{,]\s*)([a-zA-Z_][a-zA-Z0-9_]*)\s*:`).ReplaceAllString(jsObjectString, `$1"$2":`)
	jsObjectString = regexp.MustCompile(`'([^']+)':`).ReplaceAllString(jsObjectString, `"$1":`)
	jsObjectString = regexp.MustCompile(`,(\s*[}\]])`).ReplaceAllString(jsObjectString, "$1")

	var jsUnits map[string]map[string]float64

	err = json.Unmarshal([]byte(jsObjectString), &jsUnits)
	if err != nil {
		t.Fatalf("Failed to parse JS object string as JSON: %v\nOriginal JS Path: %s\nProcessed JS Object String:\n%s", err, jsPath, jsObjectString)
	}

	compareUnitMap(t, "length", jsUnits["length"], UnitConversionsLength)
	compareUnitMap(t, "duration", jsUnits["duration"], UnitConversionsDuration)
	compareUnitMap(t, "angle", jsUnits["angle"], UnitConversionsAngle)
}

func compareUnitMap(t *testing.T, category string, jsMap map[string]float64, goMap map[string]float64) {
	t.Helper()

	for jsKey, jsValue := range jsMap {
		goValue, exists := goMap[jsKey]
		if !exists {
			t.Errorf("[%s] Unit '%s' exists in JS but not in Go", category, jsKey)
			continue
		}
		if !floatsAlmostEqual(goValue, jsValue) {
			t.Errorf("[%s] Unit '%s' has different values: JS=%.15f, Go=%.15f", category, jsKey, jsValue, goValue)
		}
	}

	for goKey, goValue := range goMap {
		jsValue, exists := jsMap[goKey]
		if !exists {
			t.Errorf("[%s] Unit '%s' exists in Go but not in JS", category, goKey)
			continue
		}
		if !floatsAlmostEqual(goValue, jsValue) {
			t.Errorf("[%s] Unit '%s' has different values (Go->JS check): Go=%.15f, JS=%.15f", category, goKey, goValue, jsValue)
		}
	}
}