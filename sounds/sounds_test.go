package sounds

import (
	"regexp"
	"testing"
)

func TestParseCategory(t *testing.T) {
	tables := []struct {
		arg string
		cat *Category
		err bool
	}{
		{
			arg: "C = p t k",
			cat: &Category{
				values: []string{"p", "t", "k"},
				Name:   "C",
			},
			err: false,
		},
		{
			// this test should fail because the category `P` is
			// already defined in the test environment
			arg: "P = p b m w",
			cat: nil,
			err: true,
		},
		{
			arg: "W = {P} q w",
			cat: &Category{
				values: []string{"p", "b", "f", "v", "m", "q", "w"},
				Name:   "W",
			},
			err: false,
		},
		{
			// this test should fail because the element `ŋ` is
			// repeated
			arg: "K = ŋ k g x ɣ ŋ",
			cat: nil,
			err: true,
		},
	}
	rl := NewRuleList()
	rl.Categories["P"] = &Category{values: []string{"p", "b", "f", "v", "m"}}
	for _, tab := range tables {
		cat, err := rl.parseCategory(tab.arg)
		switch {
		case tab.err && err == nil:
			t.Errorf("parseCategory(%#v) failed to produce an error", tab.arg)
		case !tab.err && err != nil:
			t.Errorf("parseCategory(%#v) incorrectly produced the error `%v`", tab.arg, err)
		case !tab.err && err == nil:
			if !tab.cat.Equal(cat) {
				t.Errorf("parseCategory(%#v) produced"+
					"the category %#v instead of %#v",
					tab.arg, cat, tab.cat)
			}
		}
	}
}

func TestParseRule(t *testing.T) {
	tables := []struct {
		arg  string
		rule *Rule
		err  bool
	}{
		{
			arg:  "a > b",
			rule: &Rule{From: "a", To: "b"},
			err:  false,
		},
		{
			arg:  "a > b / c_",
			rule: &Rule{From: "a", To: "b", Before: "c"},
			err:  false,
		},
		{
			arg:  "a > b / c",
			rule: nil,
			err:  true,
		},
		{
			arg:  "a > b / c_d",
			rule: &Rule{From: "a", To: "b", Before: "c", After: "d"},
			err:  false,
		},
		{
			arg:  "a > b ! e_f",
			rule: &Rule{From: "a", To: "b", UnBefore: "e", UnAfter: "f"},
			err:  false,
		},
		{
			arg:  "a > b / _d",
			rule: &Rule{From: "a", To: "b", After: "d"},
			err:  false,
		},
	}
	rl := NewRuleList()
	for _, tab := range tables {
		rule, err := rl.parseRule(tab.arg)
		switch {
		case tab.err && err == nil:
			t.Errorf("parseRule(%#v) failed to produce an error", tab.arg)
		case !tab.err && err != nil:
			t.Errorf("parseRule(%#v) incorrectly produced the error %#v", tab.arg, err)
		case !tab.err && err == nil:
			if *tab.rule != *rule {
				t.Errorf("parseRule(%#v) produced the rule %#v instead of %#v",
					tab.arg, rule, tab.rule)
			}
		}
	}
}

func TestCompileRule(t *testing.T) {
	tables := []struct {
		rule *Rule
		cr   *CompiledRule
		err  bool
	}{
		{
			rule: &Rule{From: "a", To: "b"},
			cr: &CompiledRule{
				From:   &compiledPattern{regexp.MustCompile("a"), []numCat{}, CategoryList{}},
				To:     "b",
				Before: &compiledPattern{regexp.MustCompile("$"), []numCat{}, CategoryList{}},
				After:  &compiledPattern{regexp.MustCompile("^"), []numCat{}, CategoryList{}},
			},
			err: false,
		},
	}
	for _, tab := range tables {
		cr, err := tab.rule.Compile(CategoryList{})
		switch {
		case tab.err && err == nil:
			t.Errorf("%v.Compile(CategoryList{}) failed to produce an error", tab.rule)
		case !tab.err && err != nil:
			t.Errorf("%v.Compile(CategoryList{}) incorrectly produced the error %#v", tab.rule, err)
		case !tab.err && err == nil:
			if !tab.cr.Equal(cr) {
				t.Errorf("%v.Compile(CategoryList{}) produced the rule %v instead of %v",
					tab.rule, cr, tab.cr)
			}
		}
	}
}

func TestFindMatches(t *testing.T) {
	tables := []struct {
		rule    string
		word    string
		matches []Match
	}{
		{
			rule: "a > e",
			word: "banana",
			matches: []Match{
				{Start: 1, End: 2, Indices: map[int]int{}},
				{Start: 3, End: 4, Indices: map[int]int{}},
				{Start: 5, End: 6, Indices: map[int]int{}},
			},
		},
		{
			rule: "a > e / n_n",
			word: "nanananan",
			matches: []Match{
				{Start: 1, End: 2, Indices: map[int]int{}},
				{Start: 3, End: 4, Indices: map[int]int{}},
				{Start: 5, End: 6, Indices: map[int]int{}},
				{Start: 7, End: 8, Indices: map[int]int{}},
			},
		},
		{
			rule: "{0:P} > {0:N}",
			word: "ta",
			matches: []Match{
				{Start: 0, End: 1, Indices: map[int]int{0: 1}},
			},
		},
		{
			rule: "{0:N} > 0 / _{0:P}",
			word: "mtnt",
			matches: []Match{
				{Start: 2, End: 3, Indices: map[int]int{0: 1}},
			},
		},
		{
			rule: "a > 0 ! _{0:P}{0:P}",
			word: "app akt",
			matches: []Match{
				{Start: 4, End: 5, Indices: map[int]int{}},
			},
		},
	}
	rl := NewRuleList()
	rl.ParseRuleCat("P = p t k")
	rl.ParseRuleCat("N = m n ŋ")
	for _, tab := range tables {
		rule, err := rl.parseRule(tab.rule)
		if err != nil {
			t.Errorf("RuleList.parseRule(%#v) incorrectly produced the error %#v", tab.rule, err)
			continue
		}
		cr, err := rule.Compile(rl.Categories)
		if err != nil {
			t.Errorf("Rule.Compile(%#v) incorrectly produced the error %#v", tab.rule, err)
			continue
		}
		matches := cr.FindMatches(tab.word)
		for i, m := range tab.matches {
			if i < len(matches) && matches[i].Equal(m) {
				continue
			}
			t.Errorf("CompiledRule(%#v).FindMatches(%#v) failed to find match number %#v, %#v", tab.rule, tab.word, i, m)
		}
		for i, m := range matches {
			if i < len(tab.matches) && tab.matches[i].Equal(m) {
				continue
			}
			t.Errorf("CompiledRule(%#v).FindMatches(%#v) incorrectly found match number %#v, %#v", tab.rule, tab.word, i, m)
		}
	}
}
