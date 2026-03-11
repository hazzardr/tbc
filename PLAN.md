# TUI Layer — Feasibility & Implementation Plan

## Summary

Adding a TUI layer is highly feasible. The sim engine is already cleanly separated from its delivery mechanisms. The core API (`sim/core/api.go`) exposes plain Go functions that take and return protobuf structs — the web server is just a thin HTTP wrapper around them. A TUI would be an equally thin wrapper: a new `sim/tui/` package that imports `sim/core` directly, with no WASM or HTTP involved.

## Why the Architecture Works

The entire public API surface needed is already in `sim/core/api.go`:

```go
core.RunRaidSim(request *proto.RaidSimRequest) *proto.RaidSimResult
core.RunRaidSimAsync(request *proto.RaidSimRequest, progress chan *proto.ProgressMetrics)
core.StatWeights(...)
core.ComputeStats(...)
```

Each spec also has a `presets.go` (e.g. `sim/druid/balance/presets.go`) that defines complete, valid `proto.RaidSimRequest` configurations in Go — these serve as ready-made starting points for a TUI without needing to build a full gear picker immediately.

## What Needs to Be Built

### 1. Input Layer — constructing `proto.RaidSimRequest`
This is the hardest part. A full request includes: class/spec, talents, gear (item IDs + gem IDs + enchant IDs), buffs, consumables, encounter settings, and rotation options.

**Recommended approach:** start with preset-only selection (pick a spec, pick a gear set, pick a rotation), then layer in customization.

### 2. Results Display Layer — rendering `proto.RaidSimResult`
The result proto is well-structured: per-player DPS, per-spell hit/crit/miss counts, aura uptimes, resource metrics, and iteration duration. A table view is a natural fit.

### 3. Progress Reporting
`RunRaidSimAsync` already sends `*proto.ProgressMetrics` to a channel. A TUI progress bar maps directly onto this.

## Recommended TUI Library

**[`bubbletea`](https://github.com/charmbracelet/bubbletea)** — idiomatic modern Go TUI using an Elm-style update model. Companion libraries:
- `bubbles` — pre-built components (lists, tables, spinners, progress bars, text inputs)
- `lipgloss` — styling and layout

No C dependencies; trivially added to `go.mod`.

**Alternative:** [`tview`](https://github.com/rivo/tview) — traditional widget-based layout (forms, tables, modals). More familiar for config-heavy UIs, less composable.

## Scope Breakdown

| Component | Effort | Notes |
|---|---|---|
| Skeleton `sim/tui/main.go` | Trivial | `sim.RegisterAll()` + call `core.RunRaidSim` with a preset, print DPS |
| Spec/gear preset picker | Small | List of specs, list of named gear sets per spec |
| Rotation/options picker | Small | Per-spec rotation options already defined in presets |
| Buffs & debuffs configuration | Small | Toggle list |
| Encounter configuration | Small | Duration, num targets, health-based option |
| Progress bar during sim | Small | Read from `ProgressMetrics` channel |
| Results display (DPS table, spell breakdown) | Small–Medium | Tabular output with per-spell metrics |
| Full interactive gear picker | Large | Hundreds of items per slot; needs fuzzy search |
| Talent picker | Medium | Tree UI is non-trivial; could start with preset-only |

## Recommended Phased Approach

### Phase 1 — Preset Runner (1–2 days)
A minimal TUI that:
- Lists available specs
- For the selected spec, lists preset gear sets and rotation options
- Lets the user configure buffs, debuffs, and encounter duration
- Runs the sim with a progress bar
- Displays a results table (DPS, top spells by damage, hit/crit/miss)

This is entirely self-contained in a new `sim/tui/` package and requires no changes to existing code.

### Phase 2 — Customization (1 week)
- Editable encounter settings (target count, duration, health-based)
- Buff/debuff toggles
- Consumables picker
- Export/import of configurations as JSON (the proto already supports this)

### Phase 3 — Full Gear & Talent Picker (1–2 weeks)
- Searchable item list per slot (backed by `core.GetGearList()`)
- Gem and enchant selection
- Talent point allocation UI

## The One Real Challenge

**Gear selection.** The item database has hundreds of entries per slot. The web UI handles this with a searchable picker backed by the full `GetGearList` result. Options for the TUI:
- **Preset-only** (Phase 1): no picker needed, just named gear sets.
- **File-based**: load an equipment spec from a JSON file at runtime (the proto JSON format is already supported via `items.EquipmentSpecFromJsonString`).
- **Interactive picker**: fuzzy-searchable list per slot (Phase 3).

## File Structure

```
sim/tui/
  main.go          # entry point, RegisterAll(), top-level bubbletea app
  model.go         # top-level app state and Update/View
  specs.go         # spec selection screen
  presets.go       # gear/rotation preset selection screen
  config.go        # buffs, debuffs, encounter config screen
  runner.go        # sim execution + progress reporting
  results.go       # results display screen
```

## Build Integration

Add a new makefile target:

```makefile
tui:
    go build -o wowsimtbc-tui ./sim/tui/main.go
```
