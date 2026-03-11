package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode"
)

var (
	postmanVarPattern      = regexp.MustCompile(`{{\s*([A-Za-z0-9_]+)\s*}}`)
	expectedStatusPattern  = regexp.MustCompile(`(?i)expected[^0-9]*(\d{3})`)
	defaultRequestTimeout  = 30 * time.Second
	defaultStepDelay       = 150 * time.Millisecond
	defaultMinimumExecuted = 5
)

type postmanRunConfig struct {
	ServiceName string
	EnvPrefix   string

	CollectionPath string
	RequiredVars   []string
	DefaultVars    map[string]string
	MinExecuted    int

	// Optional map: if request name contains key, capture root `id` into value.
	IDCaptureHints map[string]string
}

type postmanCollection struct {
	Variable []postmanVariable `json:"variable"`
	Item     []postmanItem     `json:"item"`
}

type postmanVariable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type postmanItem struct {
	Name    string        `json:"name"`
	Item    []postmanItem `json:"item"`
	Request *postmanReq   `json:"request"`
}

type postmanReq struct {
	Method string          `json:"method"`
	Header []postmanHeader `json:"header"`
	URL    any             `json:"url"`
	Body   *postmanBody    `json:"body"`
}

type postmanHeader struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Disabled bool   `json:"disabled"`
}

type postmanBody struct {
	Mode string `json:"mode"`
	Raw  string `json:"raw"`
}

type requestStep struct {
	Name   string
	Method string
	URL    string
	Header []postmanHeader
	Body   string
}

func runPostmanCollectionE2E(t *testing.T, cfg postmanRunConfig) {
	t.Helper()
	if !envBool(cfg.EnvPrefix+"RUN", false) && !envBool("RUN_E2E", false) {
		t.Skipf("%s E2E disabled; set %sRUN=true (or RUN_E2E=true)", cfg.ServiceName, cfg.EnvPrefix)
	}

	collectionPath := cfg.CollectionPath
	if !filepath.IsAbs(collectionPath) {
		collectionPath = filepath.Clean(collectionPath)
	}

	raw, err := os.ReadFile(collectionPath)
	if err != nil {
		t.Fatalf("read collection %s: %v", collectionPath, err)
	}

	var collection postmanCollection
	if err := json.Unmarshal(raw, &collection); err != nil {
		t.Fatalf("unmarshal collection %s: %v", collectionPath, err)
	}

	vars := loadVariables(collection.Variable, raw, cfg.EnvPrefix)
	applyDefaultVars(vars, cfg.DefaultVars)
	missing := missingVars(vars, cfg.RequiredVars)
	if len(missing) > 0 {
		t.Fatalf("missing required variables for %s: %v", cfg.ServiceName, missing)
	}

	steps := flattenPostmanItems(collection.Item)
	if len(steps) == 0 {
		t.Fatalf("no executable requests found in %s", collectionPath)
	}

	timeout := envDuration(cfg.EnvPrefix+"REQUEST_TIMEOUT", defaultRequestTimeout)
	client := &http.Client{Timeout: timeout}
	stepDelay := envDuration(cfg.EnvPrefix+"STEP_DELAY", defaultStepDelay)
	strict := envBool(cfg.EnvPrefix+"STRICT", true)

	executed := 0
	failed := 0
	skipped := 0
	skippedReasons := map[string]int{}

	for _, step := range steps {
		resolvedURL, unresolvedURL := substituteVars(step.URL, vars)
		if strings.HasPrefix(resolvedURL, "ws://") || strings.HasPrefix(resolvedURL, "wss://") {
			skipped++
			skippedReasons["websocket"]++
			continue
		}

		headers := make(http.Header)
		unresolved := make([]string, 0, len(unresolvedURL))
		unresolved = append(unresolved, unresolvedURL...)

		for _, h := range step.Header {
			if h.Disabled {
				continue
			}
			key := strings.TrimSpace(h.Key)
			if key == "" {
				continue
			}
			value, missingInValue := substituteVars(h.Value, vars)
			unresolved = append(unresolved, missingInValue...)
			headers.Add(key, value)
		}

		body := ""
		if strings.TrimSpace(step.Body) != "" {
			var bodyMissing []string
			body, bodyMissing = substituteVars(step.Body, vars)
			unresolved = append(unresolved, bodyMissing...)
		}

		unresolved = dedupeStrings(unresolved)
		if len(unresolved) > 0 {
			if strict {
				failed++
				t.Logf("FAIL %s %s (%s): unresolved vars %v", step.Method, resolvedURL, step.Name, unresolved)
				continue
			}
			skipped++
			skippedReasons["missing_vars"]++
			t.Logf("SKIP %s %s (%s): unresolved vars %v", step.Method, resolvedURL, step.Name, unresolved)
			continue
		}

		req, err := http.NewRequest(step.Method, resolvedURL, bytes.NewBufferString(body))
		if err != nil {
			failed++
			t.Logf("FAIL %s (%s): build request: %v", step.Name, resolvedURL, err)
			continue
		}
		req.Header = headers
		if req.Header.Get("Content-Type") == "" && body != "" {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := client.Do(req)
		if err != nil {
			failed++
			t.Logf("FAIL %s %s (%s): %v", step.Method, resolvedURL, step.Name, err)
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		executed++

		if !statusMatches(step.Name, resp.StatusCode) {
			failed++
			t.Logf("FAIL %s %s (%s): status=%d body=%s", step.Method, resolvedURL, step.Name, resp.StatusCode, shrink(string(respBody), 400))
		} else {
			captureVarsFromResponse(step.Name, vars, respBody, cfg.IDCaptureHints)
		}

		if stepDelay > 0 {
			time.Sleep(stepDelay)
		}
	}

	minExecuted := cfg.MinExecuted
	if minExecuted <= 0 {
		minExecuted = defaultMinimumExecuted
	}
	if executed < minExecuted {
		t.Fatalf("%s E2E executed only %d requests (min=%d). skipped=%d reasons=%v", cfg.ServiceName, executed, minExecuted, skipped, skippedReasons)
	}
	if failed > 0 {
		t.Fatalf("%s E2E failed: executed=%d failed=%d skipped=%d reasons=%v", cfg.ServiceName, executed, failed, skipped, skippedReasons)
	}
	t.Logf("%s E2E passed: executed=%d skipped=%d reasons=%v", cfg.ServiceName, executed, skipped, skippedReasons)
}

func loadVariables(collectionVars []postmanVariable, collectionRaw []byte, envPrefix string) map[string]string {
	out := make(map[string]string, len(collectionVars)+8)
	for _, item := range collectionVars {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			continue
		}
		out[key] = item.Value
	}
	for _, match := range postmanVarPattern.FindAllStringSubmatch(string(collectionRaw), -1) {
		if len(match) != 2 {
			continue
		}
		key := strings.TrimSpace(match[1])
		if key == "" {
			continue
		}
		if _, exists := out[key]; !exists {
			out[key] = ""
		}
	}
	for key := range out {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			out[key] = v
		}
		if v := strings.TrimSpace(os.Getenv(envPrefix + key)); v != "" {
			out[key] = v
		}
	}
	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}
		if !strings.HasPrefix(parts[0], envPrefix) {
			continue
		}
		key := strings.TrimPrefix(parts[0], envPrefix)
		if strings.TrimSpace(key) == "" {
			continue
		}
		out[key] = strings.TrimSpace(parts[1])
	}
	return out
}

func missingVars(vars map[string]string, required []string) []string {
	missing := make([]string, 0, len(required))
	for _, key := range required {
		if strings.TrimSpace(vars[key]) == "" {
			missing = append(missing, key)
		}
	}
	return missing
}

func applyDefaultVars(vars map[string]string, defaults map[string]string) {
	for key, defaultValue := range defaults {
		if strings.TrimSpace(vars[key]) != "" {
			continue
		}
		resolved, _ := substituteVars(defaultValue, vars)
		vars[key] = strings.TrimSpace(resolved)
	}
}

func flattenPostmanItems(items []postmanItem) []requestStep {
	steps := make([]requestStep, 0, 64)
	var walk func(prefix string, items []postmanItem)
	walk = func(prefix string, items []postmanItem) {
		for _, item := range items {
			name := strings.TrimSpace(item.Name)
			fullName := name
			if prefix != "" {
				fullName = prefix + " / " + name
			}
			if len(item.Item) > 0 {
				walk(fullName, item.Item)
				continue
			}
			if item.Request == nil {
				continue
			}
			body := ""
			if item.Request.Body != nil && strings.EqualFold(item.Request.Body.Mode, "raw") {
				body = item.Request.Body.Raw
			}
			steps = append(steps, requestStep{
				Name:   fullName,
				Method: strings.ToUpper(strings.TrimSpace(item.Request.Method)),
				URL:    requestURL(item.Request.URL),
				Header: item.Request.Header,
				Body:   body,
			})
		}
	}
	walk("", items)
	return steps
}

func requestURL(raw any) string {
	switch typed := raw.(type) {
	case string:
		return strings.TrimSpace(typed)
	case map[string]any:
		if v, ok := typed["raw"].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
		host := ""
		if v, ok := typed["host"].([]any); ok {
			parts := make([]string, 0, len(v))
			for _, p := range v {
				if s, ok := p.(string); ok {
					parts = append(parts, s)
				}
			}
			host = strings.Join(parts, "")
		}
		path := ""
		if v, ok := typed["path"].([]any); ok {
			parts := make([]string, 0, len(v))
			for _, p := range v {
				if s, ok := p.(string); ok {
					parts = append(parts, s)
				}
			}
			path = "/" + strings.Join(parts, "/")
		}
		return host + path
	default:
		return ""
	}
}

func substituteVars(input string, vars map[string]string) (string, []string) {
	if strings.TrimSpace(input) == "" {
		return input, nil
	}
	missing := make([]string, 0)
	out := postmanVarPattern.ReplaceAllStringFunc(input, func(match string) string {
		keyMatch := postmanVarPattern.FindStringSubmatch(match)
		if len(keyMatch) != 2 {
			return match
		}
		key := keyMatch[1]
		value, exists := vars[key]
		if !exists {
			missing = append(missing, key)
			return match
		}
		return strings.TrimSpace(value)
	})
	return out, dedupeStrings(missing)
}

func statusMatches(stepName string, statusCode int) bool {
	if expected, ok := expectedStatus(stepName); ok {
		return statusCode == expected
	}
	return statusCode >= 200 && statusCode < 300
}

func expectedStatus(stepName string) (int, bool) {
	matches := expectedStatusPattern.FindStringSubmatch(stepName)
	if len(matches) != 2 {
		return 0, false
	}
	code, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, false
	}
	return code, true
}

func captureVarsFromResponse(stepName string, vars map[string]string, body []byte, hints map[string]string) {
	if len(body) == 0 {
		return
	}
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return
	}
	walkAndCapture(vars, payload)

	for contains, key := range hints {
		if !strings.Contains(strings.ToLower(stepName), strings.ToLower(contains)) {
			continue
		}
		if strings.TrimSpace(vars[key]) != "" {
			return
		}
		root, ok := payload.(map[string]any)
		if !ok {
			return
		}
		idValue, ok := root["id"].(string)
		if ok && strings.TrimSpace(idValue) != "" {
			vars[key] = strings.TrimSpace(idValue)
		}
		return
	}
}

func walkAndCapture(vars map[string]string, value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, inner := range typed {
			if s, ok := inner.(string); ok && strings.TrimSpace(s) != "" {
				varName := toEnvVarName(key)
				if strings.HasSuffix(varName, "_ID") && strings.TrimSpace(vars[varName]) == "" {
					vars[varName] = strings.TrimSpace(s)
				}
				if _, exists := vars[varName]; exists && strings.TrimSpace(vars[varName]) == "" {
					vars[varName] = strings.TrimSpace(s)
				}
			}
			walkAndCapture(vars, inner)
		}
	case []any:
		for _, inner := range typed {
			walkAndCapture(vars, inner)
		}
	}
}

func toEnvVarName(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	var b strings.Builder
	for i, r := range raw {
		if unicode.IsUpper(r) && i > 0 {
			b.WriteByte('_')
		}
		if r == '-' || r == '.' || r == ' ' {
			b.WriteByte('_')
			continue
		}
		b.WriteRune(unicode.ToUpper(r))
	}
	return b.String()
}

func dedupeStrings(items []string) []string {
	if len(items) <= 1 {
		return items
	}
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func envBool(key string, fallback bool) bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if raw == "" {
		return fallback
	}
	switch raw {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func envDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return parsed
}

func shrink(raw string, max int) string {
	raw = strings.TrimSpace(raw)
	if len(raw) <= max {
		return raw
	}
	return fmt.Sprintf("%s...(truncated)", raw[:max])
}
