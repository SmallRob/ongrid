package ruleengine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// RuleEngine is the top-level entry point. It wraps a RuleSet and
// provides loading, mutation, and matching operations.
type RuleEngine struct {
	rs *RuleSet
}

// NewRuleEngine returns an engine with an empty rule set.
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{rs: newRuleSet()}
}

// rulesFile is the YAML/JSON envelope: a top-level "rules" array.
type rulesFile struct {
	Rules []Rule `json:"rules" yaml:"rules"`
}

// LoadFromYAML reads a YAML file containing a "rules" array and merges
// them into the engine. Existing rules with the same ID are replaced.
func (e *RuleEngine) LoadFromYAML(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ruleengine: read yaml %q: %w", path, err)
	}
	var rf rulesFile
	if err := yaml.Unmarshal(data, &rf); err != nil {
		return fmt.Errorf("ruleengine: parse yaml %q: %w", path, err)
	}
	e.rs.mu.Lock()
	defer e.rs.mu.Unlock()
	for _, r := range rf.Rules {
		if !r.Enabled {
			continue
		}
		e.rs.put(r)
	}
	return nil
}

// LoadFromJSON reads a JSON file containing a "rules" array and merges
// them into the engine. Existing rules with the same ID are replaced.
func (e *RuleEngine) LoadFromJSON(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ruleengine: read json %q: %w", path, err)
	}
	var rf rulesFile
	if err := json.Unmarshal(data, &rf); err != nil {
		return fmt.Errorf("ruleengine: parse json %q: %w", path, err)
	}
	e.rs.mu.Lock()
	defer e.rs.mu.Unlock()
	for _, r := range rf.Rules {
		if !r.Enabled {
			continue
		}
		e.rs.put(r)
	}
	return nil
}

// LoadFromDir walks dir and loads every .yaml, .yml, and .json file.
// Files that fail to parse are skipped with a logged warning; the
// engine continues loading the rest.
func (e *RuleEngine) LoadFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("ruleengine: read dir %q: %w", dir, err)
	}
	var firstErr error
	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		full := filepath.Join(dir, ent.Name())
		ext := strings.ToLower(filepath.Ext(ent.Name()))
		var loadErr error
		switch ext {
		case ".yaml", ".yml":
			loadErr = e.LoadFromYAML(full)
		case ".json":
			loadErr = e.LoadFromJSON(full)
		default:
			continue
		}
		if loadErr != nil && firstErr == nil {
			firstErr = fmt.Errorf("ruleengine: load %q: %w", full, loadErr)
		}
	}
	return firstErr
}

// AddRule inserts or replaces a single rule in the engine.
func (e *RuleEngine) AddRule(rule Rule) {
	e.rs.mu.Lock()
	defer e.rs.mu.Unlock()
	e.rs.put(rule)
}

// RemoveRule removes a rule by ID. Returns true if the rule existed.
func (e *RuleEngine) RemoveRule(id string) bool {
	e.rs.mu.Lock()
	defer e.rs.mu.Unlock()
	return e.rs.delete(id)
}

// Rules returns a snapshot of all enabled rules.
func (e *RuleEngine) Rules() []Rule {
	e.rs.mu.RLock()
	defer e.rs.mu.RUnlock()
	return e.rs.all()
}

// Match returns all rules that match the query, sorted by descending
// score (then descending Priority as a tiebreaker). Rules with a zero
// score are omitted.
func (e *RuleEngine) Match(query string) []RuleMatch {
	return e.MatchWithContext(query, nil)
}

// MatchWithContext works like Match but also evaluates the rule's
// Conditions against the supplied context map. A rule whose conditions
// are not satisfied is skipped even if its patterns match.
func (e *RuleEngine) MatchWithContext(query string, context map[string]string) []RuleMatch {
	e.rs.mu.RLock()
	rules := e.rs.all()
	e.rs.mu.RUnlock()

	var matches []RuleMatch
	for _, r := range rules {
		if !matchConditions(r, context) {
			continue
		}
		score, patterns := e.matchScore(query, r)
		if score > 0 {
			matches = append(matches, RuleMatch{
				Rule:            r,
				Score:           score,
				MatchedPatterns: patterns,
			})
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Score != matches[j].Score {
			return matches[i].Score > matches[j].Score
		}
		return matches[i].Rule.Priority > matches[j].Rule.Priority
	})
	return matches
}

// matchScore computes a score for how well a rule matches the query.
// Each matched keyword pattern adds 10 points; each matched regex adds
// 20 points (regexes are more specific). Returns the total score and
// the list of patterns that matched.
func (e *RuleEngine) matchScore(query string, rule Rule) (int, []string) {
	lowerQuery := strings.ToLower(query)
	var score int
	var matched []string
	seen := make(map[string]bool)

	for _, pat := range rule.Patterns {
		if seen[pat] {
			continue
		}
		if strings.HasPrefix(pat, "re:") {
			pattern := strings.TrimPrefix(pat, "re:")
			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}
			if re.MatchString(query) {
				score += 20
				matched = append(matched, pat)
				seen[pat] = true
			}
		} else {
			if strings.Contains(lowerQuery, strings.ToLower(pat)) {
				score += 10
				matched = append(matched, pat)
				seen[pat] = true
			}
		}
	}
	return score, matched
}

// matchConditions checks that every key-value pair in the rule's
// Conditions map is present in the supplied context. Nil or empty
// conditions always pass.
func matchConditions(rule Rule, context map[string]string) bool {
	if len(rule.Conditions) == 0 {
		return true
	}
	for k, v := range rule.Conditions {
		if ctx, ok := context[k]; !ok || ctx != v {
			return false
		}
	}
	return true
}
