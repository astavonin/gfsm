// run-new-generator.go
//
// Usage: go generate
// In your source file, include a directive such as:
//
//	//go:generate gfsm_uml -format=plantuml
//
// This generator reads the file specified by the GOFILE environment variable,
// then for each state machine builder chain (identified by a terminating Build()
// call), it extracts the SM name from a SetSMName call and collects all
// RegisterState transitions. It then writes a diagram for each state machine
// into a separate file.
package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

const (
	setSMNameCall     = "SetSMName"
	registerStateCall = "RegisterState"
	buildCall         = "Build"
)

// Transition represents a state transition.
type Transition struct {
	Source       string
	Destinations []string
}

// StateMachine holds the name and all transitions for a state machine.
type StateMachine struct {
	Name        string
	Transitions []Transition
}

func main() {
	format := getFormat()
	filename, err := getOutputName()
	if err != nil {
		log.Fatalf("Failed to get output file name: %v", err)
	}

	machines, err := doParse(filename)
	if err != nil {
		log.Fatalf("Parse error: %s", err)
	}

	err = writeDiagram(machines, format, filename)
	if err != nil {
		log.Fatalf("Failed to write diagram: %v", err)
	}
}

func doParse(filename string) (map[string]StateMachine, error) {
	// Parse the file.
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %q: %v", filename, err)
	}

	// Map to hold state machines keyed by their name.
	machines := make(map[string]StateMachine)

	// Walk the AST to find builder chains ending with a call to Build().
	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Look for calls to Build() which terminate a builder chain.
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != buildCall {
			return true
		}

		// Extract the full call chain.
		chain := extractCallChain(call)

		// Look for the custom naming function and register state calls.
		smName, transitions := processChain(chain)
		if smName == "" {
			// If no SM name is provided, you might skip or assign a default name.
			smName = "Unnamed"
		}

		// Merge with any previously discovered machine of the same name.
		if existing, ok := machines[smName]; ok {
			existing.Transitions = append(existing.Transitions, transitions...)
			machines[smName] = existing
		} else {
			machines[smName] = StateMachine{
				Name:        smName,
				Transitions: transitions,
			}
		}

		return true
	})
	return machines, nil
}

func getOutputName() (string, error) {
	// Use `GOFILE` environment variable (set by go generate).
	filename := os.Getenv("GOFILE")
	if filename == "" {
		return "", errors.New("GOFILE environment variable is not set")
	}
	return filename, nil
}

func getFormat() string {
	// Define flag for the output format ("mermaid" or "plantuml").
	var format string
	flag.StringVar(&format, "format", "mermaid", "output format: mermaid or plantuml")
	flag.Parse()
	return format
}

func writeDiagram(machines map[string]StateMachine, format string, filename string) error {
	// Write each state machine's diagram to a file.
	for _, sm := range machines {
		var output string
		outFmt := strings.ToLower(format)
		switch outFmt {
		case "mermaid":
			output = buildMermaid(sm)
		case "plantuml":
			output = buildPlantUML(sm)
		default:
			return fmt.Errorf("unknown output format: %s", outFmt)
		}
		ext := ".mermaid"
		if outFmt == "plantuml" {
			ext = ".uml"
		}
		outFilename := sm.Name + ext
		err := os.WriteFile(outFilename, []byte(output), 0644)
		if err != nil {
			return fmt.Errorf("failed to write diagram to file %q: %v", outFilename, err)
		}
		log.Printf("Diagram for state machine %q written to %s\n\n", sm.Name, outFilename)
	}

	// If no state machines were found, log a message.
	if len(machines) == 0 {
		log.Println("No state machine definitions found in", filename)
	}
	return nil
}

// extractCallChain traverses the fluent API call chain starting at expr.
// It returns a slice of CallExpr pointers in the order they were invoked.
func extractCallChain(expr ast.Expr) []*ast.CallExpr {
	var chain []*ast.CallExpr
	current := expr
	for {
		call, ok := current.(*ast.CallExpr)
		if !ok {
			break
		}
		chain = append(chain, call)
		// Each call is a selector, e.g. previousCall.Method()
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			break
		}
		current = sel.X
	}
	// Reverse the chain so that it is in left-to-right order.
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain
}

// processChain looks through the call chain for a SetSMName call and RegisterState calls.
// It returns the SM name (from SetSMName) and a slice of transitions.
func processChain(chain []*ast.CallExpr) (string, []Transition) {
	var smName string
	var transitions []Transition

	// Iterate over each call in the chain.
	for _, callExpr := range chain {
		sel, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		methodName := sel.Sel.Name

		switch methodName {
		case setSMNameCall:
			// Expect a single argument: a string literal with the SM name.
			if len(callExpr.Args) >= 1 {
				if lit, ok := callExpr.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					smName = strings.Trim(lit.Value, `"`)
				}
			}
		case registerStateCall:
			// Expect: RegisterState(source, stateInstance, []SM{dest1, dest2, ...})
			if len(callExpr.Args) < 3 {
				continue
			}
			// First argument: source state (identifier).
			srcIdent, ok := callExpr.Args[0].(*ast.Ident)
			if !ok {
				continue
			}
			source := srcIdent.Name

			// Third argument: allowed transitions as a composite literal.
			compLit, ok := callExpr.Args[2].(*ast.CompositeLit)
			if !ok {
				continue
			}
			var dests []string
			for _, elt := range compLit.Elts {
				if ident, ok := elt.(*ast.Ident); ok {
					dests = append(dests, ident.Name)
				}
			}
			transitions = append(transitions, Transition{
				Source:       source,
				Destinations: dests,
			})
		}
	}
	return smName, transitions
}

// buildMermaid generates a Mermaid state diagram for the state machine.
func buildMermaid(sm StateMachine) string {
	var b strings.Builder
	b.WriteString("```mermaid\n")
	b.WriteString("stateDiagram-v2\n")
	for _, t := range sm.Transitions {
		for _, dest := range t.Destinations {
			b.WriteString(fmt.Sprintf("    %s --> %s\n", t.Source, dest))
		}
	}
	b.WriteString("```\n")
	return b.String()
}

// buildPlantUML generates a PlantUML state diagram for the state machine.
func buildPlantUML(sm StateMachine) string {
	var b strings.Builder
	b.WriteString("@startuml\n")
	for _, t := range sm.Transitions {
		for _, dest := range t.Destinations {
			b.WriteString(fmt.Sprintf("%s --> %s\n", t.Source, dest))
		}
	}
	b.WriteString("@enduml\n")
	return b.String()
}
