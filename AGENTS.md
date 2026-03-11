# AGENTS.md

This document provides essential information for coding agents operating within this repository to maintain consistency, quality, and functionality.

## Project Overview
This repository contains:
- **Backend:** Go-based simulation engine located in `sim/` and `generate_items/`.
- **Frontend:** TypeScript-based UI located in `ui/`.
- **Build System:** Driven by a `makefile` at the root.

## Build, Test, and Lint Commands

### Building
- **Full Build (Frontend UI):** `make` (This builds all UI components into the `dist/tbc` directory).
- **Backend Server:** `make wowsimtbc` (Builds the web server).
- **WASM Module:** `make wasm` (Compiles Go code to WebAssembly for the frontend simulation).
- **Dev Server:** `make rundevserver` (Starts the backend server locally).

### Testing
- **Run All Tests:** `make test` (Executes `go test ./...`).
- **Run Single Test (Go):** 
  To run tests in a specific package, use `go test -v ./path/to/package`.
  Example: `go test -v ./sim/druid/balance`.
  To run a specific test function: `go test -v ./sim/druid/balance -run TestBalanceName`.
- **Update Test Results:** `make update-tests` (If simulation output changes, run this to update expected results).

### Linting and Formatting
- **Full Format (Go + TS):** `make fmt`
- **Go Formatting:** `gofmt -w ./sim ./generate_items`
- **TypeScript Formatting:** `make tsfmt` (Uses `tsfmt.json` configuration).

## Code Style Guidelines

### Go (Backend)
- **Formatting:** Always run `make fmt` before committing.
- **Naming:** Follow Go idiomatic conventions (CamelCase, concise but descriptive names).
- **Structure:**
  - Logic for simulations resides in `sim/core/` and specific folders per class/spec.
  - Protobuf definitions are in `proto/`. Do NOT manually edit generated Go files in `sim/core/proto/`.
- **Error Handling:** Use standard Go error handling patterns.
- **Dependencies:** Avoid unnecessary external dependencies.

### TypeScript (Frontend)
- **Formatting:** Adhere to `tsfmt.json`. Run `make tsfmt` frequently.
- **Typing:** Strict typing is expected. Use defined interfaces and types, avoid `any`.
- **Imports:** Prefer absolute imports where possible within the project structure, maintaining consistency with existing code.
- **Frameworks:** Uses `protobuf-ts` for communication with the Go backend.
- **CSS:** Uses SCSS for styling (`_sim.scss`, `index.scss`). Follow existing module-based styling patterns.

### General Rules
- **Do not modify build configuration files** (`makefile`, `package.json`, `tsfmt.json`) unless explicitly requested.
- **Respect the directory structure.** Do not move files between directories without ensuring imports and the `makefile` (if necessary) are updated.
- **Safety:** Always verify changes by running relevant tests. If unsure about a change's impact on simulations, run the tests for the affected class/spec.
- **Pre-commit:** The repository uses a pre-commit hook (`pre-commit`) to enforce formatting. Ensure it is active (`make setup`).

## Best Practices
- When adding a new simulation feature, ensure you add the corresponding tests in the `*_test.go` file for that spec.
- When creating a new UI component, follow the structure in `ui/core/components/`.
- If you encounter a complex bug, create a reproduction test case first.
