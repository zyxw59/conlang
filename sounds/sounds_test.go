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
		// This test shouldn't fail anymore
		/* {
			// this test should fail because the element `ŋ` is
			// repeated
			arg: "K = ŋ k g x ɣ ŋ",
			cat: nil,
			err: true,
		}, */
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
				From: &compiledPattern{
					Regexp: regexp.MustCompile("a"),
				},
				To: "b",
				Before: &compiledPattern{
					Regexp: regexp.MustCompile("(?:)$"),
				},
				After: &compiledPattern{
					Regexp: regexp.MustCompile("^(?:)"),
				},
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
			rule: "a > e / N_N",
			word: "NaNaNaNaN",
			matches: []Match{
				{Start: 1, End: 2, Indices: map[int]int{}},
				{Start: 3, End: 4, Indices: map[int]int{}},
				{Start: 5, End: 6, Indices: map[int]int{}},
				{Start: 7, End: 8, Indices: map[int]int{}},
			},
		},
		{
			rule: "NaN > NeN",
			word: "NaNaNaNaN",
			matches: []Match{
				{Start: 0, End: 3, Indices: map[int]int{}},
				{Start: 4, End: 7, Indices: map[int]int{}},
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

func TestApply(t *testing.T) {
	tables := []struct {
		rule   string
		word   string
		output string
		err    bool
	}{
		{
			rule:   "a > e",
			word:   "banana",
			output: "benene",
			err:    false,
		},
		{
			rule:   "a > e / N_N",
			word:   "NaNaNaNaN",
			output: "NeNeNeNeN",
			err:    false,
		},
		{
			rule:   "NaN > NeN",
			word:   "NaNaNaNaN",
			output: "NeNaNeNaN",
			err:    false,
		},
		{
			rule:   "{0:P} > {0:N}",
			word:   "ta",
			output: "na",
			err:    false,
		},
		{
			rule:   "{0:N} > 0 / _{0:P}",
			word:   "mtnt",
			output: "mtt",
			err:    false,
		},
		{
			rule:   "a > 0 ! _{0:P}{0:P}",
			word:   "app akt",
			output: "app kt",
			err:    false,
		},
		{
			rule:   "0 > a / _#",
			word:   "top taco",
			output: "topa tacoa",
			err:    false,
		},
		{
			rule:   "{0:Vu} > {0:Va} / #({C}+{V1})*{C}+_({C}+{V0})*{C}*#",
			word:   "tap tapak takatə",
			output: "táp tapák takátə",
			err:    false,
		},
	}
	rl := NewRuleList()
	rl.ParseRuleCat("P = p t k")
	rl.ParseRuleCat("N = m n ŋ")
	rl.ParseRuleCat("C = {P} {N}")
	rl.ParseRuleCat("Vu = a e i o u")
	rl.ParseRuleCat("Va = á é í ó ú")
	rl.ParseRuleCat("V0 = ə")
	rl.ParseRuleCat("V1 = {Vu} {V0}")
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
		output, _, err := cr.Apply(tab.word)
		switch {
		case tab.err && err == nil:
			t.Errorf("Apply(%#v, %#v) failed to produce an error", tab.rule, tab.word)
		case !tab.err && err != nil:
			t.Errorf("Apply(%#v, %#v) incorrectly produced the error %#v", tab.rule, tab.word, err)
		case !tab.err && err == nil:
			if tab.output != output {
				t.Errorf("Apply(%#v, %#v) produced the output %#v instead of %#v", tab.rule, tab.word, output, tab.output)
				s, nc, err := rl.Categories.categoryReplace("({C}+{V1})")
				t.Logf("%v %v %v", s, nc, err)
			}
		}
	}
}

func TestApplyFile(t *testing.T) {
	tables := []struct {
		word   string
		output string
		err    bool
	}{
		{
			word:   "bp",
			output: "pp",
			err:    false,
		},
		{
			word:   "mbp",
			output: "mbp",
			err:    false,
		},
		{
			word:   "abadega",
			output: "awaeɣa",
			err:    false,
		},
	}
	filename := "test_sc"
	rl, err := LoadFile(filename)
	if err != nil {
		t.Fatalf("LoadFile(%#v) incorrectly produced the error %#v", filename, err)
	}
	for _, tab := range tables {
		output, _, err := rl.Apply(tab.word)
		switch {
		case tab.err && err == nil:
			t.Errorf("Apply(%#v) failed to produce an error", tab.word)
		case !tab.err && err != nil:
			t.Errorf("Apply(%#v) incorrectly produced the error %#v", tab.word, err)
		case !tab.err && err == nil:
			if tab.output != output {
				t.Errorf("Apply(%#v) produced the output %#v instead of %#v", tab.word, output, tab.output)
			}
		}
	}
}

func TestPairs(t *testing.T) {
	tables := []struct {
		names  []string
		output []string
		err    bool
	}{
		{
			names:  []string{"", ".a.b.c"},
			output: []string{"a", "a.b", "a.b.c"},
			err:    false,
		},
		{
			names:  []string{"", ".a.b.c", "c", ".d.e"},
			output: []string{"a", "a.b", "a.b.c", "c.d", "c.d.e"},
			err:    false,
		},
		{
			names:  []string{"", "a.b.c"},
			output: nil,
			err:    true,
		},
		{
			names:  []string{"", ".a.b.c", ""},
			output: nil,
			err:    true,
		},
	}
	for _, tab := range tables {
		output, err := Pairs(tab.names...)
		switch {
		case tab.err && err == nil:
			t.Errorf("Pairs(%v) failed to produce an error", tab.names)
		case !tab.err && err != nil:
			t.Errorf("Pairs(%v) incorrectly produced the error %#v", tab.names, err)
		case !tab.err && err == nil:
			for i, s := range tab.output {
				if i < len(output) && output[i] == s {
					continue
				}
				t.Errorf("Pairs(%v) wrongly produced %v instead of %v", tab.names, output, tab.output)
			}
		}
	}
}
