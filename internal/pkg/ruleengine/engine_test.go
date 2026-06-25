package ruleengine

import (
	"os"
	"path/filepath"
	"testing"
)

// --- helpers ---------------------------------------------------------------

func mustWriteTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file %s: %v", name, err)
	}
	return path
}

// --- LoadFromYAML ----------------------------------------------------------

func TestLoadFromYAML(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `
rules:
  - id: test-yaml-1
    title: "Test YAML Rule"
    description: "A rule loaded from YAML"
    category: test
    tags: [yaml, test]
    patterns: ["keyword1", "keyword2"]
    priority: 50
    response: "handle keyword1 or keyword2"
    conditions: {}
    enabled: true
  - id: test-yaml-2
    title: "Disabled Rule"
    description: "Should not be loaded"
    category: test
    tags: [disabled]
    patterns: ["never"]
    priority: 1
    response: "should not appear"
    conditions: {}
    enabled: false
`
	path := mustWriteTempFile(t, dir, "rules.yaml", yamlContent)

	eng := NewRuleEngine()
	if err := eng.LoadFromYAML(path); err != nil {
		t.Fatalf("LoadFromYAML: %v", err)
	}

	rules := eng.Rules()
	if len(rules) != 1 {
		t.Fatalf("expected 1 enabled rule, got %d", len(rules))
	}
	if rules[0].ID != "test-yaml-1" {
		t.Errorf("expected id test-yaml-1, got %s", rules[0].ID)
	}
}

// --- LoadFromJSON ----------------------------------------------------------

func TestLoadFromJSON(t *testing.T) {
	dir := t.TempDir()
	jsonContent := `{
  "rules": [
    {
      "id": "test-json-1",
      "title": "Test JSON Rule",
      "description": "A rule loaded from JSON",
      "category": "test",
      "tags": ["json"],
      "patterns": ["json-keyword"],
      "priority": 60,
      "response": "json response",
      "conditions": {},
      "enabled": true
    }
  ]
}`
	path := mustWriteTempFile(t, dir, "rules.json", jsonContent)

	eng := NewRuleEngine()
	if err := eng.LoadFromJSON(path); err != nil {
		t.Fatalf("LoadFromJSON: %v", err)
	}

	rules := eng.Rules()
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Title != "Test JSON Rule" {
		t.Errorf("expected title 'Test JSON Rule', got %q", rules[0].Title)
	}
}

// --- LoadFromDir -----------------------------------------------------------

func TestLoadFromDir(t *testing.T) {
	dir := t.TempDir()
	mustWriteTempFile(t, dir, "a.yaml", `
rules:
  - id: dir-a
    title: "Dir A"
    description: ""
    category: test
    tags: []
    patterns: ["alpha"]
    priority: 10
    response: "a"
    conditions: {}
    enabled: true
`)
	mustWriteTempFile(t, dir, "b.json", `{
  "rules": [
    {
      "id": "dir-b",
      "title": "Dir B",
      "description": "",
      "category": "test",
      "tags": [],
      "patterns": ["beta"],
      "priority": 20,
      "response": "b",
      "conditions": {},
      "enabled": true
    }
  ]
}`)
	// Non-rule files should be ignored.
	mustWriteTempFile(t, dir, "readme.txt", "not a rule file")

	eng := NewRuleEngine()
	if err := eng.LoadFromDir(dir); err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}

	if got := len(eng.Rules()); got != 2 {
		t.Fatalf("expected 2 rules, got %d", got)
	}
}

// --- Keyword matching ------------------------------------------------------

func TestMatch_Keyword(t *testing.T) {
	eng := NewRuleEngine()
	eng.AddRule(Rule{
		ID:       "kw-test",
		Title:    "Keyword Test",
		Patterns: []string{"disk full", "no space left"},
		Priority: 50,
		Response: "clean up disk",
		Enabled:  true,
	})

	matches := eng.Match("The disk full alert fired on node-3")
	if len(matches) == 0 {
		t.Fatal("expected at least one match")
	}
	if matches[0].Rule.ID != "kw-test" {
		t.Errorf("expected kw-test, got %s", matches[0].Rule.ID)
	}
	if matches[0].Score != 10 {
		t.Errorf("expected score 10 for single keyword, got %d", matches[0].Score)
	}
}

func TestMatch_KeywordCaseInsensitive(t *testing.T) {
	eng := NewRuleEngine()
	eng.AddRule(Rule{
		ID:       "case-test",
		Title:    "Case Test",
		Patterns: []string{"CPU"},
		Priority: 50,
		Response: "check cpu",
		Enabled:  true,
	})

	matches := eng.Match("cpu usage is high")
	if len(matches) == 0 {
		t.Fatal("expected case-insensitive match")
	}
}

func TestMatch_NoMatch(t *testing.T) {
	eng := NewRuleEngine()
	eng.AddRule(Rule{
		ID:       "no-match",
		Title:    "No Match",
		Patterns: []string{"quantum computing"},
		Priority: 50,
		Response: "n/a",
		Enabled:  true,
	})

	matches := eng.Match("disk space running low")
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

// --- Regex matching --------------------------------------------------------

func TestMatch_Regex(t *testing.T) {
	eng := NewRuleEngine()
	eng.AddRule(Rule{
		ID:       "regex-test",
		Title:    "Regex Test",
		Patterns: []string{`re:cpu.*(?:usage|util).*[89]\d%`},
		Priority: 80,
		Response: "CPU at high utilisation",
		Enabled:  true,
	})

	matches := eng.Match("cpu usage is at 95% on web-1")
	if len(matches) == 0 {
		t.Fatal("expected regex match")
	}
	if matches[0].Score != 20 {
		t.Errorf("expected score 20 for regex match, got %d", matches[0].Score)
	}
	if len(matches[0].MatchedPatterns) != 1 {
		t.Errorf("expected 1 matched pattern, got %d", len(matches[0].MatchedPatterns))
	}
}

func TestMatch_InvalidRegexSkipped(t *testing.T) {
	eng := NewRuleEngine()
	eng.AddRule(Rule{
		ID:       "bad-regex",
		Title:    "Bad Regex",
		Patterns: []string{`re:[invalid(`},
		Priority: 10,
		Response: "should never match",
		Enabled:  true,
	})

	// Should not panic; invalid regex is silently skipped.
	matches := eng.Match("anything")
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for invalid regex, got %d", len(matches))
	}
}

// --- Priority sorting ------------------------------------------------------

func TestMatch_PrioritySort(t *testing.T) {
	eng := NewRuleEngine()
	eng.AddRule(Rule{
		ID:       "low-pri",
		Title:    "Low Priority",
		Patterns: []string{"alert"},
		Priority: 10,
		Response: "low",
		Enabled:  true,
	})
	eng.AddRule(Rule{
		ID:       "high-pri",
		Title:    "High Priority",
		Patterns: []string{"alert"},
		Priority: 90,
		Response: "high",
		Enabled:  true,
	})

	matches := eng.Match("alert fired")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	// Same score (both match keyword "alert"), so higher priority wins.
	if matches[0].Rule.ID != "high-pri" {
		t.Errorf("expected high-pri first, got %s", matches[0].Rule.ID)
	}
	if matches[1].Rule.ID != "low-pri" {
		t.Errorf("expected low-pri second, got %s", matches[1].Rule.ID)
	}
}

func TestMatch_ScoreOverPriority(t *testing.T) {
	eng := NewRuleEngine()
	eng.AddRule(Rule{
		ID:       "score-high",
		Title:    "Score High",
		Patterns: []string{"disk", "full", "no space"},
		Priority: 10,
		Response: "multiple keywords",
		Enabled:  true,
	})
	eng.AddRule(Rule{
		ID:       "score-low",
		Title:    "Score Low",
		Patterns: []string{"disk"},
		Priority: 99,
		Response: "single keyword",
		Enabled:  true,
	})

	matches := eng.Match("disk full no space left")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	// Rule with more matches should rank first despite lower priority.
	if matches[0].Rule.ID != "score-high" {
		t.Errorf("expected score-high first (higher score), got %s", matches[0].Rule.ID)
	}
}

// --- Context matching (Conditions) -----------------------------------------

func TestMatchWithContext(t *testing.T) {
	eng := NewRuleEngine()
	eng.AddRule(Rule{
		ID:       "ctx-prod",
		Title:    "Production Only",
		Patterns: []string{"cpu"},
		Priority: 80,
		Response: "prod cpu alert",
		Conditions: map[string]string{
			"env": "production",
		},
		Enabled: true,
	})
	eng.AddRule(Rule{
		ID:       "ctx-any",
		Title:    "Any Env",
		Patterns: []string{"cpu"},
		Priority: 50,
		Response: "generic cpu alert",
		Conditions: map[string]string{},
		Enabled:  true,
	})

	// With production context, both should match; ctx-prod ranks higher.
	prodMatches := eng.MatchWithContext("cpu high", map[string]string{"env": "production"})
	if len(prodMatches) != 2 {
		t.Fatalf("expected 2 matches with prod ctx, got %d", len(prodMatches))
	}
	if prodMatches[0].Rule.ID != "ctx-prod" {
		t.Errorf("expected ctx-prod first with prod context, got %s", prodMatches[0].Rule.ID)
	}

	// With staging context, only ctx-any should match.
	stagingMatches := eng.MatchWithContext("cpu high", map[string]string{"env": "staging"})
	if len(stagingMatches) != 1 {
		t.Fatalf("expected 1 match with staging ctx, got %d", len(stagingMatches))
	}
	if stagingMatches[0].Rule.ID != "ctx-any" {
		t.Errorf("expected ctx-any with staging context, got %s", stagingMatches[0].Rule.ID)
	}

	// With nil context, only rules without conditions match.
	nilMatches := eng.MatchWithContext("cpu high", nil)
	if len(nilMatches) != 1 {
		t.Fatalf("expected 1 match with nil ctx, got %d", len(nilMatches))
	}
	if nilMatches[0].Rule.ID != "ctx-any" {
		t.Errorf("expected ctx-any with nil context, got %s", nilMatches[0].Rule.ID)
	}
}

// --- AddRule / RemoveRule --------------------------------------------------

func TestAddRule(t *testing.T) {
	eng := NewRuleEngine()
	if len(eng.Rules()) != 0 {
		t.Fatal("expected empty engine")
	}

	eng.AddRule(Rule{
		ID:       "add-1",
		Title:    "Added",
		Patterns: []string{"test"},
		Enabled:  true,
	})
	if len(eng.Rules()) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(eng.Rules()))
	}

	// Overwrite same ID.
	eng.AddRule(Rule{
		ID:       "add-1",
		Title:    "Updated",
		Patterns: []string{"test", "updated"},
		Enabled:  true,
	})
	rules := eng.Rules()
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule after overwrite, got %d", len(rules))
	}
	if rules[0].Title != "Updated" {
		t.Errorf("expected title Updated, got %q", rules[0].Title)
	}
}

func TestRemoveRule(t *testing.T) {
	eng := NewRuleEngine()
	eng.AddRule(Rule{
		ID:       "rm-1",
		Title:    "To Remove",
		Patterns: []string{"rm"},
		Enabled:  true,
	})

	if ok := eng.RemoveRule("rm-1"); !ok {
		t.Fatal("expected RemoveRule to return true")
	}
	if len(eng.Rules()) != 0 {
		t.Fatalf("expected 0 rules after removal, got %d", len(eng.Rules()))
	}

	// Removing non-existent rule should return false.
	if ok := eng.RemoveRule("no-such-id"); ok {
		t.Error("expected RemoveRule to return false for missing ID")
	}
}

// --- DefaultRules ----------------------------------------------------------

func TestDefaultRules(t *testing.T) {
	rules := DefaultRules()
	if len(rules) < 15 {
		t.Fatalf("expected at least 15 default rules, got %d", len(rules))
	}

	categories := make(map[string]bool)
	for _, r := range rules {
		if r.ID == "" {
			t.Error("rule with empty ID")
		}
		if r.Title == "" {
			t.Errorf("rule %s has empty title", r.ID)
		}
		if len(r.Patterns) == 0 {
			t.Errorf("rule %s has no patterns", r.ID)
		}
		if r.Response == "" {
			t.Errorf("rule %s has empty response", r.ID)
		}
		if !r.Enabled {
			t.Errorf("rule %s should be enabled", r.ID)
		}
		categories[r.Category] = true
	}

	// Verify all five categories are present.
	for _, cat := range []string{"system", "network", "database", "application", "security"} {
		if !categories[cat] {
			t.Errorf("missing expected category %q in default rules", cat)
		}
	}
}

// --- DefaultRules loaded into engine + YAML roundtrip ----------------------

func TestDefaultRulesIntoEngine(t *testing.T) {
	eng := NewRuleEngine()
	for _, r := range DefaultRules() {
		eng.AddRule(r)
	}

	if got := len(eng.Rules()); got < 15 {
		t.Fatalf("expected at least 15 rules in engine, got %d", got)
	}

	// Quick spot-check: querying "disk full" should surface sys-disk-full.
	matches := eng.Match("disk full on production server")
	if len(matches) == 0 {
		t.Fatal("expected matches for 'disk full'")
	}
	found := false
	for _, m := range matches {
		if m.Rule.ID == "sys-disk-full" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected sys-disk-full in results for 'disk full'")
	}
}

func TestLoadFromYAML_DefaultsFile(t *testing.T) {
	// Exercise the shipped YAML file if it exists relative to the test.
	path := filepath.Join("rules", "defaults.yaml")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("defaults.yaml not found at %s (skipping)", path)
	}

	eng := NewRuleEngine()
	if err := eng.LoadFromYAML(path); err != nil {
		t.Fatalf("LoadFromYAML defaults: %v", err)
	}

	rules := eng.Rules()
	if len(rules) < 15 {
		t.Fatalf("expected at least 15 rules from defaults.yaml, got %d", len(rules))
	}
}

// --- Multiple pattern match scoring ----------------------------------------

func TestMatch_MultiplePatternScore(t *testing.T) {
	eng := NewRuleEngine()
	eng.AddRule(Rule{
		ID:       "multi",
		Title:    "Multi Pattern",
		Patterns: []string{"disk", "full", "re:disk.*(?:usage|util).*9[0-9]%"},
		Priority: 50,
		Response: "disk at capacity",
		Enabled:  true,
	})

	matches := eng.Match("disk full and disk usage at 98% on /dev/sda")
	if len(matches) == 0 {
		t.Fatal("expected match")
	}
	// Should match "disk" (10) + "full" (10) + regex (20) = 40
	if matches[0].Score != 40 {
		t.Errorf("expected score 40 (2 keywords + 1 regex), got %d", matches[0].Score)
	}
	if len(matches[0].MatchedPatterns) != 3 {
		t.Errorf("expected 3 matched patterns, got %d", len(matches[0].MatchedPatterns))
	}
}
