package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/cli"
)

func TestApply(t *testing.T) {
	statePath := testTempFile(t)

	p := testProvider()
	ui := new(cli.MockUi)
	c := &ApplyCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	args := []string{
		"-state", statePath,
		testFixturePath("apply"),
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("err: %s", err)
	}

	f, err := os.Open(statePath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer f.Close()

	state, err := terraform.ReadState(f)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if state == nil {
		t.Fatal("state should not be nil")
	}
}

func TestApply_configInvalid(t *testing.T) {
	p := testProvider()
	ui := new(cli.MockUi)
	c := &ApplyCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	args := []string{
		"-state", testTempFile(t),
		testFixturePath("apply-config-invalid"),
	}
	if code := c.Run(args); code != 1 {
		t.Fatalf("bad: \n%s", ui.OutputWriter.String())
	}
}

func TestApply_defaultState(t *testing.T) {
	td, err := ioutil.TempDir("", "tf")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	statePath := filepath.Join(td, DefaultStateFilename)

	// Change to the temporary directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := os.Chdir(filepath.Dir(statePath)); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Chdir(cwd)

	p := testProvider()
	ui := new(cli.MockUi)
	c := &ApplyCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	args := []string{
		testFixturePath("apply"),
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("err: %s", err)
	}

	f, err := os.Open(statePath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer f.Close()

	state, err := terraform.ReadState(f)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if state == nil {
		t.Fatal("state should not be nil")
	}
}

func TestApply_error(t *testing.T) {
	statePath := testTempFile(t)

	p := testProvider()
	ui := new(cli.MockUi)
	c := &ApplyCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	var lock sync.Mutex
	errored := false
	p.ApplyFn = func(
		info *terraform.InstanceInfo,
		s *terraform.InstanceState,
		d *terraform.InstanceDiff) (*terraform.InstanceState, error) {
		lock.Lock()
		defer lock.Unlock()

		if !errored {
			errored = true
			return nil, fmt.Errorf("error")
		}

		return &terraform.InstanceState{ID: "foo"}, nil
	}
	p.DiffFn = func(
		*terraform.InstanceInfo,
		*terraform.InstanceState,
		*terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
		return &terraform.InstanceDiff{
			Attributes: map[string]*terraform.ResourceAttrDiff{
				"ami": &terraform.ResourceAttrDiff{
					New: "bar",
				},
			},
		}, nil
	}

	args := []string{
		"-state", statePath,
		testFixturePath("apply-error"),
	}
	if code := c.Run(args); code != 1 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("err: %s", err)
	}

	f, err := os.Open(statePath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer f.Close()

	state, err := terraform.ReadState(f)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if state == nil {
		t.Fatal("state should not be nil")
	}
	if len(state.RootModule().Resources) == 0 {
		t.Fatal("no resources in state")
	}
}

func TestApply_noArgs(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := os.Chdir(testFixturePath("plan")); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Chdir(cwd)

	statePath := testTempFile(t)

	p := testProvider()
	ui := new(cli.MockUi)
	c := &ApplyCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	args := []string{
		"-state", statePath,
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("err: %s", err)
	}

	f, err := os.Open(statePath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer f.Close()

	state, err := terraform.ReadState(f)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if state == nil {
		t.Fatal("state should not be nil")
	}
}

func TestApply_plan(t *testing.T) {
	planPath := testPlanFile(t, &terraform.Plan{
		Config: new(config.Config),
	})
	statePath := testTempFile(t)

	p := testProvider()
	ui := new(cli.MockUi)
	c := &ApplyCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	args := []string{
		"-state", statePath,
		planPath,
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("err: %s", err)
	}

	f, err := os.Open(statePath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer f.Close()

	state, err := terraform.ReadState(f)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if state == nil {
		t.Fatal("state should not be nil")
	}
}

func TestApply_planWithVarFile(t *testing.T) {
	varFileDir := testTempDir(t)
	varFilePath := filepath.Join(varFileDir, "terraform.tfvars")
	if err := ioutil.WriteFile(varFilePath, []byte(applyVarFile), 0644); err != nil {
		t.Fatalf("err: %s", err)
	}

	planPath := testPlanFile(t, &terraform.Plan{
		Config: new(config.Config),
	})
	statePath := testTempFile(t)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := os.Chdir(varFileDir); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Chdir(cwd)

	p := testProvider()
	ui := new(cli.MockUi)
	c := &ApplyCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	args := []string{
		"-state", statePath,
		planPath,
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("err: %s", err)
	}

	f, err := os.Open(statePath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer f.Close()

	state, err := terraform.ReadState(f)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if state == nil {
		t.Fatal("state should not be nil")
	}
}

func TestApply_planVars(t *testing.T) {
	planPath := testPlanFile(t, &terraform.Plan{
		Config: new(config.Config),
	})
	statePath := testTempFile(t)

	p := testProvider()
	ui := new(cli.MockUi)
	c := &ApplyCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	args := []string{
		"-state", statePath,
		"-var", "foo=bar",
		planPath,
	}
	if code := c.Run(args); code == 0 {
		t.Fatal("should've failed")
	}
}

func TestApply_refresh(t *testing.T) {
	originalState := &terraform.State{
		Modules: []*terraform.ModuleState{
			&terraform.ModuleState{
				Path: []string{"root"},
				Resources: map[string]*terraform.ResourceState{
					"test_instance.foo": &terraform.ResourceState{
						Type: "test_instance",
						Primary: &terraform.InstanceState{
							ID: "bar",
						},
					},
				},
			},
		},
	}

	statePath := testStateFile(t, originalState)

	p := testProvider()
	ui := new(cli.MockUi)
	c := &ApplyCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	args := []string{
		"-state", statePath,
		testFixturePath("apply"),
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	if !p.RefreshCalled {
		t.Fatal("should call refresh")
	}

	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("err: %s", err)
	}

	f, err := os.Open(statePath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer f.Close()

	state, err := terraform.ReadState(f)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if state == nil {
		t.Fatal("state should not be nil")
	}

	// Should have a backup file
	f, err = os.Open(statePath + DefaultBackupExtention)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	backupState, err := terraform.ReadState(f)
	f.Close()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actualStr := strings.TrimSpace(backupState.String())
	expectedStr := strings.TrimSpace(originalState.String())
	if actualStr != expectedStr {
		t.Fatalf("bad:\n\n%s\n\n%s", actualStr, expectedStr)
	}
}

func TestApply_shutdown(t *testing.T) {
	stopped := false
	stopCh := make(chan struct{})
	stopReplyCh := make(chan struct{})

	statePath := testTempFile(t)

	p := testProvider()
	shutdownCh := make(chan struct{})
	ui := new(cli.MockUi)
	c := &ApplyCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},

		ShutdownCh: shutdownCh,
	}

	p.DiffFn = func(
		*terraform.InstanceInfo,
		*terraform.InstanceState,
		*terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
		return &terraform.InstanceDiff{
			Attributes: map[string]*terraform.ResourceAttrDiff{
				"ami": &terraform.ResourceAttrDiff{
					New: "bar",
				},
			},
		}, nil
	}
	p.ApplyFn = func(
		*terraform.InstanceInfo,
		*terraform.InstanceState,
		*terraform.InstanceDiff) (*terraform.InstanceState, error) {
		if !stopped {
			stopped = true
			close(stopCh)
			<-stopReplyCh
		}

		return &terraform.InstanceState{
			ID: "foo",
			Attributes: map[string]string{
				"ami": "2",
			},
		}, nil
	}

	go func() {
		<-stopCh
		shutdownCh <- struct{}{}

		// This is really dirty, but we have no other way to assure that
		// tf.Stop() has been called. This doesn't assure it either, but
		// it makes it much more certain.
		time.Sleep(50 * time.Millisecond)

		close(stopReplyCh)
	}()

	args := []string{
		"-state", statePath,
		testFixturePath("apply-shutdown"),
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("err: %s", err)
	}

	f, err := os.Open(statePath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer f.Close()

	state, err := terraform.ReadState(f)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if state == nil {
		t.Fatal("state should not be nil")
	}

	if len(state.RootModule().Resources) != 1 {
		t.Fatalf("bad: %d", len(state.RootModule().Resources))
	}
}

func TestApply_state(t *testing.T) {
	originalState := &terraform.State{
		Modules: []*terraform.ModuleState{
			&terraform.ModuleState{
				Path: []string{"root"},
				Resources: map[string]*terraform.ResourceState{
					"test_instance.foo": &terraform.ResourceState{
						Type: "test_instance",
						Primary: &terraform.InstanceState{
							ID: "bar",
						},
					},
				},
			},
		},
	}

	statePath := testStateFile(t, originalState)

	p := testProvider()
	p.DiffReturn = &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"ami": &terraform.ResourceAttrDiff{
				New: "bar",
			},
		},
	}

	ui := new(cli.MockUi)
	c := &ApplyCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	// Run the apply command pointing to our existing state
	args := []string{
		"-state", statePath,
		testFixturePath("apply"),
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	// Verify that the provider was called with the existing state
	actual := strings.TrimSpace(p.DiffState.String())
	expected := strings.TrimSpace(testApplyStateDiffStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}

	actual = strings.TrimSpace(p.ApplyState.String())
	expected = strings.TrimSpace(testApplyStateStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}

	// Verify a new state exists
	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("err: %s", err)
	}

	f, err := os.Open(statePath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer f.Close()

	state, err := terraform.ReadState(f)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if state == nil {
		t.Fatal("state should not be nil")
	}

	// Should have a backup file
	f, err = os.Open(statePath + DefaultBackupExtention)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	backupState, err := terraform.ReadState(f)
	f.Close()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// nil out the ConnInfo since that should not be restored
	originalState.RootModule().Resources["test_instance.foo"].Primary.Ephemeral.ConnInfo = nil

	actualStr := strings.TrimSpace(backupState.String())
	expectedStr := strings.TrimSpace(originalState.String())
	if actualStr != expectedStr {
		t.Fatalf("bad:\n\n%s\n\n%s", actualStr, expectedStr)
	}
}

func TestApply_stateNoExist(t *testing.T) {
	p := testProvider()
	ui := new(cli.MockUi)
	c := &ApplyCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	args := []string{
		"idontexist.tfstate",
		testFixturePath("apply"),
	}
	if code := c.Run(args); code != 1 {
		t.Fatalf("bad: \n%s", ui.OutputWriter.String())
	}
}

func TestApply_vars(t *testing.T) {
	statePath := testTempFile(t)

	p := testProvider()
	ui := new(cli.MockUi)
	c := &ApplyCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	actual := ""
	p.DiffFn = func(
		info *terraform.InstanceInfo,
		s *terraform.InstanceState,
		c *terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
		if v, ok := c.Config["value"]; ok {
			actual = v.(string)
		}

		return &terraform.InstanceDiff{}, nil
	}

	args := []string{
		"-var", "foo=bar",
		"-state", statePath,
		testFixturePath("apply-vars"),
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	if actual != "bar" {
		t.Fatal("didn't work")
	}
}

func TestApply_varFile(t *testing.T) {
	varFilePath := testTempFile(t)
	if err := ioutil.WriteFile(varFilePath, []byte(applyVarFile), 0644); err != nil {
		t.Fatalf("err: %s", err)
	}

	statePath := testTempFile(t)

	p := testProvider()
	ui := new(cli.MockUi)
	c := &ApplyCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	actual := ""
	p.DiffFn = func(
		info *terraform.InstanceInfo,
		s *terraform.InstanceState,
		c *terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
		if v, ok := c.Config["value"]; ok {
			actual = v.(string)
		}

		return &terraform.InstanceDiff{}, nil
	}

	args := []string{
		"-var-file", varFilePath,
		"-state", statePath,
		testFixturePath("apply-vars"),
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	if actual != "bar" {
		t.Fatal("didn't work")
	}
}

func TestApply_varFileDefault(t *testing.T) {
	varFileDir := testTempDir(t)
	varFilePath := filepath.Join(varFileDir, "terraform.tfvars")
	if err := ioutil.WriteFile(varFilePath, []byte(applyVarFile), 0644); err != nil {
		t.Fatalf("err: %s", err)
	}

	statePath := testTempFile(t)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := os.Chdir(varFileDir); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Chdir(cwd)

	p := testProvider()
	ui := new(cli.MockUi)
	c := &ApplyCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	actual := ""
	p.DiffFn = func(
		info *terraform.InstanceInfo,
		s *terraform.InstanceState,
		c *terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
		if v, ok := c.Config["value"]; ok {
			actual = v.(string)
		}

		return &terraform.InstanceDiff{}, nil
	}

	args := []string{
		"-state", statePath,
		testFixturePath("apply-vars"),
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	if actual != "bar" {
		t.Fatal("didn't work")
	}
}

func TestApply_backup(t *testing.T) {
	originalState := &terraform.State{
		Modules: []*terraform.ModuleState{
			&terraform.ModuleState{
				Path: []string{"root"},
				Resources: map[string]*terraform.ResourceState{
					"test_instance.foo": &terraform.ResourceState{
						Type: "test_instance",
						Primary: &terraform.InstanceState{
							ID: "bar",
						},
					},
				},
			},
		},
	}

	statePath := testStateFile(t, originalState)
	backupPath := testTempFile(t)

	p := testProvider()
	p.DiffReturn = &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"ami": &terraform.ResourceAttrDiff{
				New: "bar",
			},
		},
	}

	ui := new(cli.MockUi)
	c := &ApplyCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	// Run the apply command pointing to our existing state
	args := []string{
		"-state", statePath,
		"-backup", backupPath,
		testFixturePath("apply"),
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	// Verify a new state exists
	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("err: %s", err)
	}

	f, err := os.Open(statePath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer f.Close()

	state, err := terraform.ReadState(f)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if state == nil {
		t.Fatal("state should not be nil")
	}

	// Should have a backup file
	f, err = os.Open(backupPath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	backupState, err := terraform.ReadState(f)
	f.Close()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := backupState.RootModule().Resources["test_instance.foo"]
	expected := originalState.RootModule().Resources["test_instance.foo"]
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v %#v", actual, expected)
	}
}

func TestApply_disableBackup(t *testing.T) {
	originalState := testState()
	statePath := testStateFile(t, originalState)

	p := testProvider()
	p.DiffReturn = &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"ami": &terraform.ResourceAttrDiff{
				New: "bar",
			},
		},
	}

	ui := new(cli.MockUi)
	c := &ApplyCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	// Run the apply command pointing to our existing state
	args := []string{
		"-state", statePath,
		"-backup", "-",
		testFixturePath("apply"),
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	// Verify that the provider was called with the existing state
	actual := strings.TrimSpace(p.DiffState.String())
	expected := strings.TrimSpace(testApplyDisableBackupStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}

	actual = strings.TrimSpace(p.ApplyState.String())
	expected = strings.TrimSpace(testApplyDisableBackupStateStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}

	// Verify a new state exists
	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("err: %s", err)
	}

	f, err := os.Open(statePath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer f.Close()

	state, err := terraform.ReadState(f)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if state == nil {
		t.Fatal("state should not be nil")
	}

	// Ensure there is no backup
	_, err = os.Stat(statePath + DefaultBackupExtention)
	if err == nil || !os.IsNotExist(err) {
		t.Fatalf("backup should not exist")
	}
}

const applyVarFile = `
foo = "bar"
`

const testApplyDisableBackupStr = `
ID = bar
`

const testApplyDisableBackupStateStr = `
ID = bar
`

const testApplyStateStr = `
ID = bar
`

const testApplyStateDiffStr = `
ID = bar
`
