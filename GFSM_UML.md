# gfsm_uml

**gfsm_uml** is a code generator tool that extracts state machine definitions from your Go source code and generates corresponding state diagrams. It supports multiple state machine definitions per file using a custom naming function (`SetSMName`), and it can output diagrams in either Mermaid or PlantUML format.

## Features

- **Automatic Diagram Generation:** Scans your code for state machine builder chains and generates diagrams automatically.
- **Custom Naming:** Use `SetSMName` in your builder chain to uniquely identify each state machine.
- **Multiple Formats:** Output diagrams in either `Mermaid` or `PlantUML` format.
- **Integration with go generate:** Easily integrate with your build process using go generate.

## Installation

If you're working from within your module, you can install the tool locally. Since `gfsm_uml` is part of the `gfsm` repository (with the command-line app located in `cmd/gfsm_uml`), install it by running:

```bash
go install github.com/astavonin/gfsm/cmd/gfsm_uml@latest
```

Make sure that your `$GOPATH/bin` (or your module-aware binary install location) is in your PATH so that you can invoke `gfsm_uml` directly.

## Usage

### Step 1. Annotate Your Code

In the Go source file that contains your state machine definitions, add a go:generate directive at the top. For example:

```go
//go:generate gfsm_uml -format=plantuml
```

Then, within your builder chain, use `SetSMName` to name your state machine. For example:

```go
return NewBuilder[StartStopSM]().
    SetSMName("FooSM").               // Custom annotation for naming the state machine
    SetDefaultState(Start).
    RegisterState(Start, &StartState{}, []StartStopSM{Stop, InProgress}).
    RegisterState(Stop, &StopState{}, []StartStopSM{Start}).
    RegisterState(InProgress, &InProgressState{}, []StartStopSM{Stop}).
    Build()
```

### Step 2. Run the Generator

From the root of your project, run:

```bash
go generate ./...
```

When invoked via `go generate`, `gfsm_uml` will read the file specified by the `GOFILE` environment variable, extract the builder chain, and generate a state diagram. If you specified PlantUML as the format (with `-format=plantuml`), the diagram will be written to a file named `FooSM.uml` (based on the name provided by `SetSMName`).

### Command Line Options

The tool supports the following flag:

- **`-format`**: Specifies the output diagram format. Valid options are:
  - `mermaid` (default)
  - `plantuml`

For example, to generate a Mermaid diagram from a file:

```bash
gfsm_uml -format=mermaid
```

## Example

Consider the following builder chain in your Go source file:

```go
return NewBuilder[StartStopSM]().
    SetSMName("FooSM").
    SetDefaultState(Start).
    RegisterState(Start, &StartState{}, []StartStopSM{Stop, InProgress}).
    RegisterState(Stop, &StopState{}, []StartStopSM{Start}).
    RegisterState(InProgress, &InProgressState{}, []StartStopSM{Stop}).
    Build()
```

When you run `go generate ./...`, the tool processes the file, extracts the state machine named `"FooSM"`, and generates a diagram file named `FooSM.uml` (if using PlantUML) or `FooSM.mermaid` (if using Mermaid).

## Troubleshooting

- **No Diagram Generated:**  
  Verify that your file contains valid state machine definitions (calls to both `SetSMName` and `RegisterState`).  
  Run `go generate -x ./...` to see detailed output and debug any issues.

- **Installation Issues:**  
  Ensure that `gfsm_uml` is installed and available in your PATH. Use `go install` as described above.
