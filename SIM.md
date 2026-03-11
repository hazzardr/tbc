# Simulation Engine Technical Overview

This document provides a technical overview of the TBC simulation engine used in this repository.

## 1. Entry Point: `RunSim`
The simulation entry point is `RunSim` in `sim/core/sim.go`. The flow is:
`RunSim` → `runSim` → `NewSim` → `(presims)` → `sim.run()`.

`NewSim` initializes the `Environment` (raid + encounter), seeds the RNG, and returns a `*Simulation` struct.

## 2. Environment Setup: Three-Phase Construction
Before any iterations run, the environment goes through three phases:

| Phase | What happens |
|---|---|
| **Construct** | `NewRaid`/`NewEncounter` created; each unit gets an `Env` pointer and a `CurrentTarget`. Debuffs applied. |
| **Initialize** | Each target and player calls `initialize()`. Buffs, talents, gear bonuses, consumes applied. |
| **Finalize** | Stat dependencies resolved; `initialStats` snapshot frozen; attack/defense tables built between every attacker↔target pair (pre-computing suppression values). |

## 3. The Outer Iteration Loop: `sim.run()`
Runs `N` iterations (`SimOptions.Iterations`). Each call to `sim.runOnce()` represents one complete fight from `0s` to `Duration`.

## 4. The Core Event Loop: `sim.runOnce()`
This is an **event-driven priority queue** loop, not a fixed-timestep loop:

```go
for {
    pa := sim.pendingActions[last]   // Pop highest-priority/soonest action
    sim.advance(pa.NextActionAt - sim.CurrentTime)  // Jump time forward
    pa.OnAction(sim)                 // Execute the action
}
```

**Termination:** Exits when the next action is past `sim.Duration` (time-based), or `Encounter.DamageTaken >= EndFightAtHealth` (health-based).

**After the loop:** `CleanUp` callbacks fire, then `doneIteration` is called on the Raid, Encounter, and all units to finalize metrics.

## 5. The Priority Queue: `PendingAction`
Every future event is a `PendingAction` with:
- `NextActionAt`: when to fire.
- `Priority`: tiebreaker ordering.

**Priorities (highest processed first):**
`DOT (3) > Auto (2) > Regen (1) > GCD (0) > Low (-1)`

## 6. Time Advancement: `sim.advance()`
Advances `sim.CurrentTime`.
1.  Checks if `executePhase` should trigger (20% health/time remaining).
2.  All `Unit`s call `auraTracker.advance()`, expiring auras whose `expires <= CurrentTime`.
3.  Hardcasts that have completed fire their `OnExpire` callback.

## 7. The Agent & GCD Loop
Each player has a GCD `PendingAction` that re-schedules itself.
- When `OnGCDReady` fires, the class-specific `Agent` runs rotation logic and calls `spell.Cast()`.
- `WaitUntil` and `WaitForMana` are mechanisms for agents to idle until a condition is met.

## 8. Casting a Spell
`spell.Cast()` is a chain of composed wrappers:
`Init → Resources → Haste → GCD → Cooldown → SharedCooldown → Wait/Deliver`

- **Instant spells:** Applied synchronously during the GCD callback.
- **Cast-time spells:** Schedules a `Hardcast` pending action. When reached, `OnExpire` delivers the spell effect.
- Mana is consumed at cast *completion*.

## 9. Spell Effect Resolution
The damage calculation pipeline:
1. `BaseDamage`
2. `applyAttackerModifiers` (school multipliers, proc flags)
3. `applyResistances` (partial resist table)
4. `OutcomeApplier` (roll hit/miss/crit/dodge/parry/glance/block)
5. `applyTargetModifiers` (target debuffs, vulnerability)
6. `finalize` (write to `SpellMetrics`, fire `OnSpellHitDealt`/`Taken` callbacks)

## 10. Auras
Auras use a `auraTracker` with bucketing (`onCastCompleteAuras`, `onSpellHitDealtAuras`, etc.) for efficient callback dispatching without iterating over all inactive auras.

## 11. DoTs
A `Dot` is an `Aura` + a `PeriodicAction` (a self-rescheduling `PendingAction`). Snapshots stats at application time so that buffs during the DoT's duration do not dynamically alter existing ticks.

## 12. Metrics & Results
Aggregated into `proto.RaidSimResult` for the UI after all iterations complete.
