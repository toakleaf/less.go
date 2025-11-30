// Code generator for visitor dispatch type switches
//
// This tool generates exhaustive type switch statements for visitor pattern
// implementations, eliminating reflection overhead.
//
// Usage: go run ./cmd/gen-visitor-dispatch
//
// The generated output can be used to update visitor implementations.

package main

import (
	"fmt"
	"sort"
	"strings"
)

// All AST node types that can be visited
var nodeTypes = []string{
	"Anonymous",
	"Assignment",
	"AtRule",
	"Attribute",
	"Call",
	"Color",
	"Combinator",
	"Comment",
	"Condition",
	"Container",
	"Declaration",
	"DetachedRuleset",
	"Dimension",
	"Element",
	"Expression",
	"Extend",
	"Import",
	"JavaScript",
	"Keyword",
	"Media",
	"MixinCall",
	"MixinDefinition",
	"NamespaceValue",
	"Negative",
	"Operation",
	"Paren",
	"Property",
	"QueryInParens",
	"Quoted",
	"Ruleset",
	"Selector",
	"UnicodeDescriptor",
	"Unit",
	"URL",
	"Value",
	"Variable",
	"VariableCall",
}

// VisitorConfig defines which methods a visitor has
type VisitorConfig struct {
	Name           string
	ReceiverVar    string
	ReceiverType   string
	VisitMethods   []string // Methods like "VisitDeclaration", "VisitRuleset"
	VisitOutMethods []string // Methods like "VisitRulesetOut"
	IsReplacing    bool     // Whether the visitor replaces nodes
}

var visitors = []VisitorConfig{
	{
		Name:         "JoinSelectorVisitor",
		ReceiverVar:  "jsv",
		ReceiverType: "*JoinSelectorVisitor",
		VisitMethods: []string{
			"Declaration",
			"MixinDefinition",
			"Ruleset",
			"Media",
			"Container",
			"AtRule",
		},
		VisitOutMethods: []string{
			"Ruleset",
		},
		IsReplacing: false,
	},
	{
		Name:         "ExtendFinderVisitor",
		ReceiverVar:  "efv",
		ReceiverType: "*ExtendFinderVisitor",
		VisitMethods: []string{
			"Declaration",
			"MixinDefinition",
			"Ruleset",
			"Media",
			"AtRule",
		},
		VisitOutMethods: []string{
			"Ruleset",
			"Media",
			"AtRule",
		},
		IsReplacing: false,
	},
	{
		Name:         "ProcessExtendsVisitor",
		ReceiverVar:  "pev",
		ReceiverType: "*ProcessExtendsVisitor",
		VisitMethods: []string{
			"Declaration",
			"MixinDefinition",
			"Selector",
			"Ruleset",
			"Media",
			"AtRule",
		},
		VisitOutMethods: []string{
			"Media",
			"AtRule",
		},
		IsReplacing: true,
	},
	{
		Name:         "ToCSSVisitor",
		ReceiverVar:  "v",
		ReceiverType: "*ToCSSVisitor",
		VisitMethods: []string{
			"Declaration",
			"MixinDefinition",
			"Extend",
			"Comment",
			"Media",
			"Container",
			"Import",
			"Anonymous",
			"Ruleset",
			"AtRule",
		},
		VisitOutMethods: []string{},
		IsReplacing:     true,
	},
	{
		Name:         "ImportVisitor",
		ReceiverVar:  "iv",
		ReceiverType: "*ImportVisitor",
		VisitMethods: []string{
			"Import",
			"Declaration",
			"AtRule",
			"MixinDefinition",
			"Ruleset",
			"Media",
		},
		VisitOutMethods: []string{
			"Declaration",
			"AtRule",
			"MixinDefinition",
			"Ruleset",
			"Media",
		},
		IsReplacing: false,
	},
}

func main() {
	fmt.Println("// Generated visitor dispatch code")
	fmt.Println("// Run: go run ./cmd/gen-visitor-dispatch")
	fmt.Println()

	for _, v := range visitors {
		generateVisitor(v)
		fmt.Println()
	}
}

func generateVisitor(config VisitorConfig) {
	fmt.Printf("// ==================== %s ====================\n\n", config.Name)

	// Generate VisitNode method
	generateVisitNode(config)
	fmt.Println()

	// Generate VisitNodeOut method
	generateVisitNodeOut(config)
}

func generateVisitNode(config VisitorConfig) {
	fmt.Printf("// VisitNode implements DirectDispatchVisitor for %s\n", config.Name)
	fmt.Printf("func (%s %s) VisitNode(node any, visitArgs *VisitArgs) (any, bool) {\n", config.ReceiverVar, config.ReceiverType)
	fmt.Printf("\tswitch n := node.(type) {\n")

	// Create a set of methods for quick lookup
	methodSet := make(map[string]bool)
	for _, m := range config.VisitMethods {
		methodSet[m] = true
	}

	// Sort node types for consistent output
	sortedTypes := make([]string, len(nodeTypes))
	copy(sortedTypes, nodeTypes)
	sort.Strings(sortedTypes)

	for _, nodeType := range sortedTypes {
		if methodSet[nodeType] {
			// This node type has a specific visit method
			fmt.Printf("\tcase *%s:\n", nodeType)
			if config.IsReplacing {
				fmt.Printf("\t\treturn %s.Visit%s(n, visitArgs), true\n", config.ReceiverVar, nodeType)
			} else {
				fmt.Printf("\t\t%s.Visit%s(n, visitArgs)\n", config.ReceiverVar, nodeType)
				fmt.Printf("\t\treturn n, true\n")
			}
		}
	}

	// Default case - return the node unhandled so visitor framework uses default behavior
	fmt.Printf("\tdefault:\n")
	fmt.Printf("\t\treturn node, true // Node type handled (no-op)\n")
	fmt.Printf("\t}\n")
	fmt.Printf("}\n")
}

func generateVisitNodeOut(config VisitorConfig) {
	fmt.Printf("// VisitNodeOut implements DirectDispatchVisitor for %s\n", config.Name)
	fmt.Printf("func (%s %s) VisitNodeOut(node any) bool {\n", config.ReceiverVar, config.ReceiverType)

	if len(config.VisitOutMethods) == 0 {
		fmt.Printf("\treturn true // No VisitOut methods, handled as no-op\n")
		fmt.Printf("}\n")
		return
	}

	fmt.Printf("\tswitch n := node.(type) {\n")

	// Create a set of methods for quick lookup
	methodSet := make(map[string]bool)
	for _, m := range config.VisitOutMethods {
		methodSet[m] = true
	}

	// Sort node types for consistent output
	sortedTypes := make([]string, len(nodeTypes))
	copy(sortedTypes, nodeTypes)
	sort.Strings(sortedTypes)

	for _, nodeType := range sortedTypes {
		if methodSet[nodeType] {
			// This node type has a specific visitOut method
			fmt.Printf("\tcase *%s:\n", nodeType)
			fmt.Printf("\t\t%s.Visit%sOut(n)\n", config.ReceiverVar, nodeType)
			fmt.Printf("\t\treturn true\n")
		}
	}

	// Default case - handled as no-op
	// Use _ to avoid "declared and not used" error
	fmt.Printf("\tdefault:\n")
	fmt.Printf("\t\t_ = n\n")
	fmt.Printf("\t\treturn true // Node type handled (no-op)\n")
	fmt.Printf("\t}\n")
	fmt.Printf("}\n")
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, s) {
			return true
		}
	}
	return false
}
