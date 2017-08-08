package sounds

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	commentstr = "//"
	arrowstr   = " > "
	equalstr   = " = "
	ruleFromTo = `(\S*) > (\S*)`
	ruleEnv    = `(?: \/ ([^\s_]*)_([^\s_]*))?`
	ruleUnEnv  = `(?: ! ([^\s_]*)_([^\s_]*))?`
)

var ruleRegExp = regexp.MustCompile(`^` + ruleFromTo + ruleEnv + ruleUnEnv + `$`)

// A RuleList is an object representing a list of sound change rules and sound
// categories used in those rules
type RuleList struct {
	Categories CategoryList
	Rules      []*Rule
	Lines      []string
}

// NewRuleList initializes an empty RuleList
func NewRuleList() *RuleList {
	return &RuleList{Categories: make(CategoryList)}
}

// A Rule is a sound change rule that changes a sound or set of sounds to
// another, in a given environment
type Rule struct {
	From     string
	To       string
	Before   string
	After    string
	UnBefore string
	UnAfter  string
}

// Equal compares two Rules by value
func (r *Rule) Equal(other *Rule) bool {
	return *r == *other
}

// ParseRuleCat takes a line and parses it as a rule or a category, adding it
// to the RuleList
func (rl *RuleList) ParseRuleCat(line string) error {
	line = strings.TrimSpace(line)
	switch {
	case len(line) == 0:
		// empty line, do nothing
	case strings.HasPrefix(line, commentstr):
		// Don't parse, it's a comment
		rl.Lines = append(rl.Lines, line)
	case strings.Contains(line, arrowstr):
		r, err := rl.parseRule(line)
		if err != nil {
			return err
		}
		rl.Rules = append(rl.Rules, r)
		rl.Lines = append(rl.Lines, line)
	case strings.Contains(line, equalstr):
		cat, err := rl.parseCategory(line)
		if err != nil {
			return err
		}
		rl.Categories[cat.Name] = cat
		rl.Lines = append(rl.Lines, line)
	default:
		return fmt.Errorf("parse error: `%s` is not a valid rule or category", line)
	}
	return nil
}

// parseRule parses a line as a rule
func (rl *RuleList) parseRule(line string) (*Rule, error) {
	matches := ruleRegExp.FindStringSubmatch(line)
	if len(matches) < 7 {
		return nil, fmt.Errorf("parse error: `%s` is not a valid rule", line)
	}
	rule := &Rule{
		From:     matches[1],
		To:       matches[2],
		Before:   matches[3],
		After:    matches[4],
		UnBefore: matches[5],
		UnAfter:  matches[6],
	}
	return rule, nil
}

// parseCategory parses a line as a category
func (rl *RuleList) parseCategory(line string) (*Category, error) {
	split := strings.SplitN(line, equalstr, 2)
	key := split[0]
	if val, ok := rl.Categories[key]; ok {
		return nil, fmt.Errorf("category error: category '%s' already defined as %v", key, val)
	}
	values := split[1]
	for k, v := range rl.Categories {
		values = strings.Replace(values, "{"+k+"}", v.String(), -1)
	}
	cat, err := NewCategory(key, strings.Fields(values))
	if err != nil {
		return nil, err
	}
	return cat, nil
}
