package terraform

import (
	"sync/atomic"
)

// stopHook is a private Hook implementation that Terraform uses to
// signal when to stop or cancel actions.
type stopHook struct {
	stop uint32
}

func (h *stopHook) PreApply(string, *InstanceState, *InstanceDiff) (HookAction, error) {
	return h.hook()
}

func (h *stopHook) PostApply(string, *InstanceState, error) (HookAction, error) {
	return h.hook()
}

func (h *stopHook) PreDiff(string, *InstanceState) (HookAction, error) {
	return h.hook()
}

func (h *stopHook) PostDiff(string, *InstanceDiff) (HookAction, error) {
	return h.hook()
}

func (h *stopHook) PreProvisionResource(string, *InstanceState) (HookAction, error) {
	return h.hook()
}

func (h *stopHook) PostProvisionResource(string, *InstanceState) (HookAction, error) {
	return h.hook()
}

func (h *stopHook) PreProvision(string, string) (HookAction, error) {
	return h.hook()
}

func (h *stopHook) PostProvision(string, string) (HookAction, error) {
	return h.hook()
}

func (h *stopHook) PreRefresh(string, *InstanceState) (HookAction, error) {
	return h.hook()
}

func (h *stopHook) PostRefresh(string, *InstanceState) (HookAction, error) {
	return h.hook()
}

func (h *stopHook) hook() (HookAction, error) {
	if h.Stopped() {
		return HookActionHalt, nil
	}

	return HookActionContinue, nil
}

// reset should be called within the lock context
func (h *stopHook) Reset() {
	atomic.StoreUint32(&h.stop, 0)
}

func (h *stopHook) Stop() {
	atomic.StoreUint32(&h.stop, 1)
}

func (h *stopHook) Stopped() bool {
	return atomic.LoadUint32(&h.stop) == 1
}
