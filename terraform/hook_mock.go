package terraform

// MockHook is an implementation of Hook that can be used for tests.
// It records all of its function calls.
type MockHook struct {
	PreApplyCalled bool
	PreApplyId     string
	PreApplyDiff   *InstanceDiff
	PreApplyState  *InstanceState
	PreApplyReturn HookAction
	PreApplyError  error

	PostApplyCalled      bool
	PostApplyId          string
	PostApplyState       *InstanceState
	PostApplyError       error
	PostApplyReturn      HookAction
	PostApplyReturnError error

	PreDiffCalled bool
	PreDiffId     string
	PreDiffState  *InstanceState
	PreDiffReturn HookAction
	PreDiffError  error

	PostDiffCalled bool
	PostDiffId     string
	PostDiffDiff   *InstanceDiff
	PostDiffReturn HookAction
	PostDiffError  error

	PreProvisionResourceCalled bool
	PreProvisionResourceId     string
	PreProvisionInstanceState  *InstanceState
	PreProvisionResourceReturn HookAction
	PreProvisionResourceError  error

	PostProvisionResourceCalled bool
	PostProvisionResourceId     string
	PostProvisionInstanceState  *InstanceState
	PostProvisionResourceReturn HookAction
	PostProvisionResourceError  error

	PreProvisionCalled        bool
	PreProvisionId            string
	PreProvisionProvisionerId string
	PreProvisionReturn        HookAction
	PreProvisionError         error

	PostProvisionCalled        bool
	PostProvisionId            string
	PostProvisionProvisionerId string
	PostProvisionReturn        HookAction
	PostProvisionError         error

	PostRefreshCalled bool
	PostRefreshId     string
	PostRefreshState  *InstanceState
	PostRefreshReturn HookAction
	PostRefreshError  error

	PreRefreshCalled bool
	PreRefreshId     string
	PreRefreshState  *InstanceState
	PreRefreshReturn HookAction
	PreRefreshError  error
}

func (h *MockHook) PreApply(n string, s *InstanceState, d *InstanceDiff) (HookAction, error) {
	h.PreApplyCalled = true
	h.PreApplyId = n
	h.PreApplyDiff = d
	h.PreApplyState = s
	return h.PreApplyReturn, h.PreApplyError
}

func (h *MockHook) PostApply(n string, s *InstanceState, e error) (HookAction, error) {
	h.PostApplyCalled = true
	h.PostApplyId = n
	h.PostApplyState = s
	h.PostApplyError = e
	return h.PostApplyReturn, h.PostApplyReturnError
}

func (h *MockHook) PreDiff(n string, s *InstanceState) (HookAction, error) {
	h.PreDiffCalled = true
	h.PreDiffId = n
	h.PreDiffState = s
	return h.PreDiffReturn, h.PreDiffError
}

func (h *MockHook) PostDiff(n string, d *InstanceDiff) (HookAction, error) {
	h.PostDiffCalled = true
	h.PostDiffId = n
	h.PostDiffDiff = d
	return h.PostDiffReturn, h.PostDiffError
}

func (h *MockHook) PreProvisionResource(id string, s *InstanceState) (HookAction, error) {
	h.PreProvisionResourceCalled = true
	h.PreProvisionResourceId = id
	h.PreProvisionInstanceState = s
	return h.PreProvisionResourceReturn, h.PreProvisionResourceError
}

func (h *MockHook) PostProvisionResource(id string, s *InstanceState) (HookAction, error) {
	h.PostProvisionResourceCalled = true
	h.PostProvisionResourceId = id
	h.PostProvisionInstanceState = s
	return h.PostProvisionResourceReturn, h.PostProvisionResourceError
}

func (h *MockHook) PreProvision(id, provId string) (HookAction, error) {
	h.PreProvisionCalled = true
	h.PreProvisionId = id
	h.PreProvisionProvisionerId = provId
	return h.PreProvisionReturn, h.PreProvisionError
}

func (h *MockHook) PostProvision(id, provId string) (HookAction, error) {
	h.PostProvisionCalled = true
	h.PostProvisionId = id
	h.PostProvisionProvisionerId = provId
	return h.PostProvisionReturn, h.PostProvisionError
}

func (h *MockHook) PreRefresh(n string, s *InstanceState) (HookAction, error) {
	h.PreRefreshCalled = true
	h.PreRefreshId = n
	h.PreRefreshState = s
	return h.PreRefreshReturn, h.PreRefreshError
}

func (h *MockHook) PostRefresh(n string, s *InstanceState) (HookAction, error) {
	h.PostRefreshCalled = true
	h.PostRefreshId = n
	h.PostRefreshState = s
	return h.PostRefreshReturn, h.PostRefreshError
}
