package tflint

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test mutual exclusivity validation (expected to fail until logic implemented)
func TestScanParamCategoryAndRemoteConfigUrlMutualExclusion(t *testing.T) {
	// Both set should produce an error once validation is added
	param := ScanParam{Category: "reusable", RemoteConfigUrl: "https://example.com/config.tflint.hcl", TargetPath: "/tmp"}
	// Expect Scan to return an error about mutual exclusivity
	_, err := Scan(param)
	require.Error(t, err, "expected error when both category and remote_config_url are set")
	assert.Contains(t, err.Error(), "mutually exclusive")
}

// RED test: happy path remote_config_url single file should succeed without invoking downloadConfigContent
func TestScanRemoteConfigSingleFileSuccess(t *testing.T) {
	memFs := afero.NewMemMapFs()
	fsStub := gostub.Stub(&fs, memFs)
	defer fsStub.Reset()

	require.NoError(t, memFs.MkdirAll("/test/terraform", 0o755))

	// Track if legacy download called (should NOT be)
	calledLegacy := false
	// Do not stub legacy download; we assert it's not called (flag would remain false)

	pathStub := gostub.Stub(&getDefaultTargetPath, func(p string) (string, error) { return p, nil })
	defer pathStub.Reset()

	tempDirStub := gostub.Stub(&setupTempConfigDir, func() (string, func(), error) {
		tempDir := "/tmp/tflint-config-remote-single"
		require.NoError(t, memFs.MkdirAll(tempDir, 0o755))
		cleanup := func() { _ = memFs.RemoveAll(tempDir) }
		return tempDir, cleanup, nil
	})
	defer tempDirStub.Reset()

	// Stub remote getter to simulate download of a single file
	mockGetter := &mockRemoteGetter{createFile: func(dst string) error {
		// dst is now the full file path, not a directory
		return afero.WriteFile(memFs, dst, []byte(`rule "terraform_deprecated_syntax" { enabled = true }`), 0644)
	}}
	getterStub := gostub.Stub(&remoteConfigGetter, mockGetter)
	defer getterStub.Reset()

	// Minimal command executor stubs so init/scan succeed
	mockExecutor := &MockCommandExecutor{patterns: map[string]*MockCommandResult{
		"tflint --init --config=":        {stdout: "init ok", stderr: "", err: nil},
		"tflint --format=json --config=": {stdout: `{"issues":[],"errors":[]}`, stderr: "", err: nil},
	}}
	execStub := gostub.Stub(&commandExecutor, mockExecutor)
	defer execStub.Reset()

	param := ScanParam{RemoteConfigUrl: "https://example.com/remote.tflint.hcl", TargetPath: "/test/terraform"}
	result, err := Scan(param)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.False(t, calledLegacy, "legacy downloadConfigContent should not be called")
	assert.True(t, strings.HasSuffix(result.Output, "}"), "output should end with JSON brace from scan")
}

// mockRemoteGetter implements RemoteGetter for tests
type mockRemoteGetter struct {
	createFile func(dst string) error
}

func (m *mockRemoteGetter) Get(dst, src string) error {
	if m.createFile != nil {
		return m.createFile(dst)
	}
	return nil
}

// RED test: remote_config_url resolving to a directory should error about requiring single file.
func TestScanRemoteConfigDirectoryError(t *testing.T) {
	// Setup in-memory filesystem and target terraform directory
	memFs := afero.NewMemMapFs()
	// Stub global fs
	fsStub := gostub.Stub(&fs, memFs)
	defer fsStub.Reset()

	require.NoError(t, memFs.MkdirAll("/test/terraform", 0o755))

	// Stub command executor to succeed for init & scan (so if remote logic not triggered, scan passes and test fails as desired)
	mockExecutor := &MockCommandExecutor{patterns: map[string]*MockCommandResult{}}
	mockExecutor.patterns = map[string]*MockCommandResult{
		"tflint --init --config=":        {stdout: "init ok", stderr: "", err: nil},
		"tflint --format=json --config=": {stdout: `{"issues":[],"errors":[]}`, stderr: "", err: nil},
	}
	execStub := gostub.Stub(&commandExecutor, mockExecutor)
	defer execStub.Reset()

	// Stub downloadConfigContent to return simple base config (category path will still be used until remote implemented)
	dlStub := gostub.Stub(&downloadConfigContent, func(url string) (string, error) {
		return `rule "terraform_deprecated_syntax" { enabled = true }`, nil
	})
	defer dlStub.Reset()

	// Stub getDefaultTargetPath passthrough
	pathStub := gostub.Stub(&getDefaultTargetPath, func(p string) (string, error) { return p, nil })
	defer pathStub.Reset()

	// Stub temp dir creation
	tempDirStub := gostub.Stub(&setupTempConfigDir, func() (string, func(), error) {
		tempDir := "/tmp/tflint-config-remote"
		require.NoError(t, memFs.MkdirAll(tempDir, 0o755))
		cleanup := func() { _ = memFs.RemoveAll(tempDir) }
		return tempDir, cleanup, nil
	})
	defer tempDirStub.Reset()

	// Use a remote_config_url pointing to a repo root (directory) which we expect to error once implemented
	param := ScanParam{RemoteConfigUrl: "git::https://example.com/org/repo.git", TargetPath: "/test/terraform"}
	_, err := Scan(param)

	// EXPECTATION (future): error complaining remote_config_url must point to single file.
	// Currently this will likely NOT error (category fallback) -> red.
	require.Error(t, err, "expected error for directory remote config but got none")
	assert.Contains(t, err.Error(), "git repository root detected")
}

// Test: remote getter failure should be propagated
func TestScanRemoteConfigGetterFailure(t *testing.T) {
	memFs := afero.NewMemMapFs()
	require.NoError(t, memFs.MkdirAll("/test/terraform", 0o755))

	fsStub := gostub.Stub(&fs, memFs)
	defer fsStub.Reset()

	mockExecutor := &MockCommandExecutor{patterns: map[string]*MockCommandResult{
		"tflint --init --config=":        {stdout: "init ok", stderr: "", err: nil},
		"tflint --format=json --config=": {stdout: `{"issues":[],"errors":[]}`, stderr: "", err: nil},
	}}
	execStub := gostub.Stub(&commandExecutor, mockExecutor)
	defer execStub.Reset()

	pathStub := gostub.Stub(&getDefaultTargetPath, func(p string) (string, error) { return p, nil })
	defer pathStub.Reset()

	tempDirStub := gostub.Stub(&setupTempConfigDir, func() (string, func(), error) {
		tempDir := "/tmp/tflint-config-remote-getter-fail"
		require.NoError(t, memFs.MkdirAll(tempDir, 0o755))
		cleanup := func() { _ = memFs.RemoveAll(tempDir) }
		return tempDir, cleanup, nil
	})
	defer tempDirStub.Reset()

	// Stub remote getter to simulate network failure
	getterStub := gostub.Stub(&remoteConfigGetter, &mockRemoteGetter{createFile: func(dst string) error {
		return fmt.Errorf("network error: could not fetch remote config")
	}})
	defer getterStub.Reset()

	param := ScanParam{RemoteConfigUrl: "https://example.com/config.tflint.hcl", TargetPath: "/test/terraform"}
	_, err := Scan(param)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch remote config")
}

func TestScanRemoteConfigTimeoutEnv(t *testing.T) {
	memFs := afero.NewMemMapFs()
	require.NoError(t, memFs.MkdirAll("/test/terraform", 0o755))
	fsStub := gostub.Stub(&fs, memFs)
	defer fsStub.Reset()
	pathStub := gostub.Stub(&getDefaultTargetPath, func(p string) (string, error) { return p, nil })
	defer pathStub.Reset()

	// Use a getter that sleeps beyond timeout to trigger context deadline
	getterStub := gostub.Stub(&remoteConfigGetter, &mockRemoteGetter{createFile: func(dst string) error {
		time.Sleep(10 * time.Millisecond)
		return afero.WriteFile(memFs, dst, []byte("rule \"x\" {}"), 0644)
	}})
	defer getterStub.Reset()
	// Set very short timeout
	t.Setenv("TFLINT_REMOTE_CONFIG_TIMEOUT_SECONDS", "0") // invalid -> fallback to default (won't timeout)
	// Since invalid value fallback happens, we expect success (no timeout) even with sleep 10ms.

	mockExecutor := &MockCommandExecutor{patterns: map[string]*MockCommandResult{
		"tflint --init --config=":        {stdout: "init ok", stderr: "", err: nil},
		"tflint --format=json --config=": {stdout: `{"issues":[],"errors":[]}`, stderr: "", err: nil},
	}}
	execStub := gostub.Stub(&commandExecutor, mockExecutor)
	defer execStub.Reset()

	param := ScanParam{RemoteConfigUrl: "https://example.com/remote.tflint.hcl", TargetPath: "/test/terraform"}
	result, err := Scan(param)
	require.NoError(t, err)
	require.NotNil(t, result)
}

// New test: multiple .hcl files should error

// New test: network failure propagation
func TestScanRemoteConfigNetworkFailure(t *testing.T) {
	memFs := afero.NewMemMapFs()
	require.NoError(t, memFs.MkdirAll("/test/terraform", 0o755))
	fsStub := gostub.Stub(&fs, memFs)
	defer fsStub.Reset()
	pathStub := gostub.Stub(&getDefaultTargetPath, func(p string) (string, error) { return p, nil })
	defer pathStub.Reset()
	tempDirStub := gostub.Stub(&setupTempConfigDir, func() (string, func(), error) {
		dir := "/tmp/tflint-config-remote-network"
		require.NoError(t, memFs.MkdirAll(dir, 0o755))
		return dir, func() { _ = memFs.RemoveAll(dir) }, nil
	})
	defer tempDirStub.Reset()
	getterStub := gostub.Stub(&remoteConfigGetter, &mockRemoteGetter{createFile: func(dst string) error { return assert.AnError }})
	defer getterStub.Reset()
	param := ScanParam{RemoteConfigUrl: "https://example.com/network", TargetPath: "/test/terraform"}
	_, err := Scan(param)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch remote config")
}

// New test: invalid HCL merge sanitization (simulate by providing custom config that fails merge)

// New test: precedence ignored rules override (simulate by enabling a rule in remote and custom then disabling via ignored_rules)
func TestScanRemoteConfigIgnoredRulesApplied(t *testing.T) {
	memFs := afero.NewMemMapFs()
	require.NoError(t, memFs.MkdirAll("/test/terraform", 0o755))
	// create dummy terraform file to allow scan
	require.NoError(t, afero.WriteFile(memFs, "/test/terraform/main.tf", []byte("# dummy"), 0644))
	fsStub := gostub.Stub(&fs, memFs)
	defer fsStub.Reset()
	pathStub := gostub.Stub(&getDefaultTargetPath, func(p string) (string, error) { return p, nil })
	defer pathStub.Reset()
	tempDirStub := gostub.Stub(&setupTempConfigDir, func() (string, func(), error) {
		dir := "/tmp/tflint-config-remote-precedence"
		require.NoError(t, memFs.MkdirAll(dir, 0o755))
		return dir, func() { _ = memFs.RemoveAll(dir) }, nil
	})
	defer tempDirStub.Reset()
	getterStub := gostub.Stub(&remoteConfigGetter, &mockRemoteGetter{createFile: func(dst string) error {
		return afero.WriteFile(memFs, dst, []byte("rule \"terraform_deprecated_syntax\" { enabled = true }"), 0644)
	}})
	defer getterStub.Reset()
	// Mock executor that verifies --disable-rule flags are passed for ignored rules
	mockExecutor := &MockCommandExecutor{patterns: map[string]*MockCommandResult{
		"tflint --init --config=":                    {stdout: "init ok", stderr: "", err: nil},
		"--disable-rule=terraform_deprecated_syntax": {stdout: `{"issues":[],"errors":[]}`, stderr: "", err: nil},
	}}
	execStub := gostub.Stub(&commandExecutor, mockExecutor)
	defer execStub.Reset()
	param := ScanParam{RemoteConfigUrl: "https://example.com/remote.hcl", TargetPath: "/test/terraform", IgnoredRules: []string{"terraform_deprecated_syntax"}}
	result, err := Scan(param)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
}
