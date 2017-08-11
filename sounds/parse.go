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

type Applier interface {
	// Apply applies a sound change to a word, or makes no change, and
	// returns the new form of the word, along with debuging information
	// and any error raised
	Apply(string) (string, string, error)
}

// A RuleList is an object representing a list of sound change rules and sound
// categories used in those rules
type RuleList struct {
	Categories CategoryList
	Lines      []Applier
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

// String writes the rule as it would appear in a sound change file
func (r *Rule) String() string {
	parts := make([]string, 3)
	parts[0] = fmt.Sprintf("%s > %s", r.From, r.To)
	if len(r.Before) > 0 || len(r.After) > 0 {
		parts[1] = fmt.Sprintf(" / %s_%s", r.Before, r.After)
	}
	if len(r.UnBefore) > 0 || len(r.UnAfter) > 0 {
		parts[2] = fmt.Sprintf(" ! %s_%s", r.UnBefore, r.UnAfter)
	}
	return strings.Join(parts, "")
}

// A Comment is a line in a sound change file that begins with `//` and is
// ignored except for debugging purposes
type Comment string

func (c Comment) Apply(word string) (output, debug string, err error) {
	return word, string(c), nil
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
		rl.Lines = append(rl.Lines, Comment(line))
	case strings.Contains(line, arrowstr):
		r, err := rl.parseRule(line)
		if err != nil {
			return err
		}
		cr, err := r.Compile(rl.Categories)
		if err != nil {
			return err
		}
		rl.Lines = append(rl.Lines, cr)
	case strings.Contains(line, equalstr):
		cat, err := rl.parseCategory(line)
		if err != nil {
			return err
		}
		rl.Categories[cat.Name] = cat
		rl.Lines = append(rl.Lines, cat)
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
	key := strings.TrimSpace(split[0])
	if val, ok := rl.Categories[key]; ok {
		return nil, fmt.Errorf("category error: category '%s' already defined as %v", key, val)
	}
	values := split[1]
	for k, v := range rl.Categories {
		values = strings.Replace(values, "{"+k+"}", v.ElemString(), -1)
	}
	cat, err := NewCategory(key, strings.Fields(values))
	if err != nil {
		return nil, err
	}
	return cat, nil
}
