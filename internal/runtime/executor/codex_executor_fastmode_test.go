package executor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	sdktranslator "github.com/router-for-me/CLIProxyAPI/v6/sdk/translator"
	"github.com/tidwall/gjson"
)

func resetClaudeFastModeCache() {
	claudeFastModeCache.Lock()
	defer claudeFastModeCache.Unlock()
	claudeFastModeCache.path = ""
	claudeFastModeCache.mtime = time.Time{}
	claudeFastModeCache.enabled = false
	claudeFastModeCache.loaded = false
}

func writeClaudeSettings(t *testing.T, dir string, body string, modTime time.Time) string {
	t.Helper()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write settings.json: %v", err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("chtimes settings.json: %v", err)
	}
	return path
}

func TestApplyClaudeFastServiceTierAddsPriorityForClaudeGPT54(t *testing.T) {
	resetClaudeFastModeCache()
	dir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", dir)
	writeClaudeSettings(t, dir, `{"fastMode":true}`, time.Now().Add(-2*time.Second))

	body := applyClaudeFastServiceTier(context.Background(), []byte(`{"model":"gpt-5.4"}`), sdktranslator.FromString("claude"), "gpt-5.4")
	if got := gjson.GetBytes(body, "service_tier").String(); got != "priority" {
		t.Fatalf("service_tier = %q, want %q", got, "priority")
	}
}

func TestApplyClaudeFastServiceTierSkipsNonClaudeOrOtherModels(t *testing.T) {
	resetClaudeFastModeCache()
	dir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", dir)
	writeClaudeSettings(t, dir, `{"fastMode":true}`, time.Now().Add(-2*time.Second))

	cases := []struct {
		name      string
		from      sdktranslator.Format
		baseModel string
	}{
		{name: "non claude source", from: sdktranslator.FromString("openai-response"), baseModel: "gpt-5.4"},
		{name: "non target model", from: sdktranslator.FromString("claude"), baseModel: "gpt-5.3"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetClaudeFastModeCache()
			body := applyClaudeFastServiceTier(context.Background(), []byte(`{"model":"`+tc.baseModel+`"}`), tc.from, tc.baseModel)
			if got := gjson.GetBytes(body, "service_tier").String(); got != "" {
				t.Fatalf("service_tier = %q, want empty", got)
			}
		})
	}
}

func TestClaudeFastModeEnabledReloadsWhenSettingsFileChanges(t *testing.T) {
	resetClaudeFastModeCache()
	dir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", dir)
	path := writeClaudeSettings(t, dir, `{"fastMode":true}`, time.Now().Add(-3*time.Second))

	enabled, ok := claudeFastModeEnabled(context.Background())
	if !ok || !enabled {
		t.Fatalf("first read = (%v, %v), want (true, true)", enabled, ok)
	}

	updatedTime := time.Now().Add(3 * time.Second)
	if err := os.WriteFile(path, []byte(`{"fastMode":false}`), 0o644); err != nil {
		t.Fatalf("rewrite settings.json: %v", err)
	}
	if err := os.Chtimes(path, updatedTime, updatedTime); err != nil {
		t.Fatalf("chtimes updated settings.json: %v", err)
	}

	enabled, ok = claudeFastModeEnabled(context.Background())
	if !ok || enabled {
		t.Fatalf("second read = (%v, %v), want (false, true)", enabled, ok)
	}
}
