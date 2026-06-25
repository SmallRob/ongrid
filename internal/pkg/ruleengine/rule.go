// Package ruleengine implements a structured knowledge-base rule engine
// for AIOps pattern matching. It loads rules from YAML/JSON files or
// in-memory definitions and matches incoming queries against keyword and
// regex patterns, returning ranked results.
//
// Designed to be consumed by the knowledge bounded context and any agent
// tool that needs fast, deterministic incident-to-response mapping.
package ruleengine

import "sync"

// Rule is a single knowledge-base rule. Patterns holds a mix of plain
// keywords (matched case-insensitively) and Go regular expressions
// (prefixed with "re:" to distinguish them from literals).
type Rule struct {
	ID          string            `json:"id"          yaml:"id"`
	Title       string            `json:"title"       yaml:"title"`
	Description string            `json:"description" yaml:"description"`
	Category    string            `json:"category"    yaml:"category"`
	Tags        []string          `json:"tags"        yaml:"tags"`
	Patterns    []string          `json:"patterns"    yaml:"patterns"`
	Priority    int               `json:"priority"    yaml:"priority"`
	Response    string            `json:"response"    yaml:"response"`
	Conditions  map[string]string `json:"conditions"  yaml:"conditions"`
	Enabled     bool              `json:"enabled"     yaml:"enabled"`
}

// RuleMatch pairs a Rule with the score it achieved against a query.
// Higher scores indicate a stronger match. MatchedPatterns records which
// of the rule's patterns triggered.
type RuleMatch struct {
	Rule            Rule
	Score           int
	MatchedPatterns []string
}

// RuleSet is a concurrency-safe collection of rules keyed by ID.
type RuleSet struct {
	rules map[string]Rule
	mu    sync.RWMutex
}

// newRuleSet returns an empty RuleSet.
func newRuleSet() *RuleSet {
	return &RuleSet{rules: make(map[string]Rule)}
}

// put inserts or replaces a rule. Caller must hold the write lock (or be
// the sole goroutine).
func (rs *RuleSet) put(r Rule) {
	rs.rules[r.ID] = r
}

// delete removes a rule by ID. Returns true if the rule existed.
func (rs *RuleSet) delete(id string) bool {
	_, ok := rs.rules[id]
	if ok {
		delete(rs.rules, id)
	}
	return ok
}

// all returns a snapshot slice of every rule (no guaranteed order).
func (rs *RuleSet) all() []Rule {
	out := make([]Rule, 0, len(rs.rules))
	for _, r := range rs.rules {
		out = append(out, r)
	}
	return out
}

// count returns the number of rules in the set.
func (rs *RuleSet) count() int {
	return len(rs.rules)
}
