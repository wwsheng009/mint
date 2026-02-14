# NodeID Usage Guide

## Core Principle

**NodeID must ONLY come from Fiber.NodeID**

❌ **NEVER** convert VNode keys to NodeID using hash
✅ **ALWAYS** use Fiber.NodeID directly

---

## Valid Patterns

### ✅ Pattern 1: Direct Fiber Access (CORRECT)

```go
// In buildComputedBox
if fiber != nil {
    box.NodeID = fiber.NodeID  // ✅ Correct: Use Fiber.NodeID
}

// In Instance creation
inst.NodeID = fiber.NodeID  // ✅ Correct

// In BuildHitMapFromFiber
nodeID := fiber.NodeID  // ✅ Correct
```

### ✅ Pattern 2: NodeID Propagation (CORRECT)

```go
// buildComputedBoxWithSize
if fiber != nil {
    box.NodeID = fiber.NodeID  // ✅ Set from Fiber
} else if parent != nil && parent.NodeID != 0 {
    box.NodeID = parent.NodeID  // ✅ Inherit from parent
} else {
    box.NodeID = 0  // ✅ Use 0 to indicate "not set"
}
```

---

## ❌ Invalid Patterns

### ❌ Pattern 1: Hashing VNode Keys (WRONG)

```go
// In distributeEventToVNode
nodeID := uint64(0)
if key := vnode.Key(); key != "" {
    nodeID = event.StringToNodeID(key)  // ❌ WRONG: Hash != Fiber.NodeID
}
```

**Why wrong?**
- VNode.Key() is a user-provided string (e.g., "button-ok", "cancel-btn")
- Fiber.NodeID is a stable tree identifier (e.g., 1234567890)
- Hashing the key creates a DIFFERENT ID that breaks event routing

### ❌ Pattern 2: Multiple Hash Conversions (WRONG)

```go
// In buildHitMapFromComputedBoxes
if nodeID == 0 && box.VNode != nil {
    if key := box.VNode.Key(); key != "" {
        nodeID = event.StringToNodeID(key)  // ❌ Hash 1
    }
}
// Later in another function
if nodeID == 0 && box.VNode != nil {
    if key := box.VNode.Key(); key != "" {
        nodeID = event.StringToNodeID(key)  // ❌ Hash 2 - Could conflict!
    }
}
```

**Why wrong?**
- Two different VNodes with the same key get the same hash
- Hash collisions cause NodeID conflicts in HitMap

---

## Migration Plan

### Phase 4.5: Remove Hashing (Current)

**Problem locations:**

1. **`internal/render/declarative_node.go:1259-1262`**
   - Event distribution uses `event.StringToNodeID(vnode.Key())`
   - Should use: `fiber.NodeID` directly from Instance

2. **`runtime/compute/engine.go:1686-1691`**
   - `buildHitMapFromComputedBoxes` hashes VNode keys when NodeID is 0
   - Should use: 0 to indicate "not set from Fiber"

### Phase 5: Complete Fiber-First NodeID (Target)

**Target state:**

1. All ComputedBox nodes have NodeID set from Fiber (never 0 or hash)
2. All Instances use NodeID from Fiber
3. All HitMap entries use NodeID from Fiber
4. No `StringToNodeID()` conversion needed anywhere

**Changes needed:**

1. **`internal/render/declarative_node.go`**
   ```go
   // Change from:
   nodeID := uint64(0)
   if key := vnode.Key(); key != "" {
       nodeID = event.StringToNodeID(key)  // ❌
   }
   
   // To:
   nodeID := inst.NodeID  // ✅ Use Instance's NodeID (comes from Fiber)
   ```

2. **`runtime/compute/engine.go`**
   ```go
   // Change from:
   if nodeID == 0 && box.VNode != nil {
       if key := box.VNode.Key(); key != "" {
           nodeID = event.StringToNodeID(key)  // ❌
       }
   }
   
   // To:
   if nodeID == 0 && box.VNode != nil {
       // Log warning: VNode found without NodeID (should not happen)
       log.Warn("ComputedBox has VNode but NodeID=0 - this indicates missing Fiber.NodeID")
       // Use 0 (no hashing)
   }
   ```

3. **`internal/reconciler/begin_work.go:216`**
   ```go
   // Comment is correct, but ensure implementation follows it:
   // "HitMap stores NodeID = Fiber.Path, so instance key must use Path too!"
   ```
   
   // This means instance key MUST match Fiber.Path, not Fiber.NodeID
   // The current implementation should already be correct

---

## Testing Strategy

### Unit Test: NodeID Consistency

```go
func TestNodeIDNotHashed(t *testing.T) {
    // Create Fiber with NodeID = 1234567890
    fiber := &Fiber{NodeID: 1234567890, ...}
    
    // Create VNode with key = "test-key"
    vnode := ui.Text("test").Key("test-key")
    
    // Hash the key
    hashID := event.StringToNodeID("test-key")
    
    // Verify they are different
    if hashID == fiber.NodeID {
        t.Error("Hash should not equal Fiber.NodeID")
    }
}

func TestNodeIDPropagation(t *testing.T) {
    // Test that NodeID propagates correctly
    fiber := createFiberTree()
    
    // Check all nodes have non-zero NodeID
    verifyNodeIDs(t, fiber)
}

func verifyNodeIDs(t *testing.T, fiber *Fiber) {
    WalkFiberDepthFirst(fiber, func(f *Fiber) bool {
        if f.NodeID == 0 {
            t.Errorf("Fiber node has NodeID=0: %v", f)
        }
        return true
    })
}
```

### Integration Test: Event Routing

```go
func TestEventRoutingUsesFiberNodeID(t *testing.T) {
    // Create app with modal
    app := createTestApp()
    
    // Trigger button click
    buttonNode := findNodeByKey(app, "modal-ok-btn")
    instance := getInstanceForNode(buttonNode)
    
    // Verify instance has Fiber.NodeID, not hash
    if instance.NodeID == 0 || instance.NodeID == hashOf("modal-ok-btn") {
        t.Error("Instance should have Fiber.NodeID, not hash")
    }
    
    // Simulate click event
    event := &MouseEvent{TargetID: instance.NodeID, ...}
    app.HandleEvent(event)
    
    // Verify routing worked
}
```

---

## Common Pitfalls to Avoid

### Pitfall 1: Using VNode Key as NodeID

**Wrong:**
```go
targetID := event.StringToNodeID("my-button")
```

**Right:**
```go
targetID := instance.NodeID  // Already set from Fiber
```

### Pitfall 2: Hashing for Comparison

**Wrong:**
```go
if key == "button" {
    nodeID = hash("button")  // Wrong: Different key, same hash
}
```

**Right:**
```go
// Don't use hash for comparison
// Use Fiber.NodeID directly or maintain a separate key map
```

### Pitfall 3: Double Hashing

**Wrong:**
```go
// Function 1:
id1 := event.StringToNodeID(key)

// Function 2:
id2 := event.StringToNodeID(key)  // Same key, same hash!
```

**Right:**
```go
// Use NodeID directly, never hash keys
```

---

## Checklist for Code Review

When reviewing code, check:

- [ ] Does this code use `event.StringToNodeID()`?
  - If yes, it's probably wrong (unless for VNode keys only)
  
- [ ] Does this code pass `NodeID` from Fiber directly?
  - If yes, it's probably correct
  
- [ ] Are we mixing VNode keys with NodeIDs?
  - If yes, refactor to use NodeID only
  
- [ ] Is NodeID = 0 or hash used as fallback?
  - NodeID = 0 is OK for "not set from Fiber"
  - NodeID = hash(key) is WRONG

---

## Related Documentation

- [Fiber Architecture](./UNIFIED_FIBER_ARCHITECTURE_REFACTOR.md)
- [NodeID System](./IDENTITY_REFACTORING_PLAN.md)
- [Phase 4.5 Transition](./PHASE_4_5_TRANSITION_STATUS.md)

---

**Last Updated**: 2026-02-14
**Status**: Draft for Review
