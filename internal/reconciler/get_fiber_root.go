package reconciler

// GetFiberRootForRendering returns Fiber root for rendering pipeline
// Phase 8: Fiber to Layout Engine NodeID propagation
// This method is called by rendering pipeline after reconciliation completes
// to ensure Fiber tree is available for NodeID propagation to ComputedBox
func (r *Reconciler) GetFiberRootForRendering() *Fiber {
	return r.root
}