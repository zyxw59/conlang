package main

import (
	"bufio"
	"bytes"
	"flag"
	"github.com/zyxw59/conlang/sounds"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"text/template"
)

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var err error

	prefix := flag.String("p", "", "prefix for sound change files")
	flag.Parse()
	filename := flag.Arg(0)

	var rl *sounds.RuleList
	cache := sounds.NewCache()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	t := template.New(filepath.Base(filename))
	t = t.Funcs(template.FuncMap{
		"ApplyPairs": func(word string, names ...string) (string, error) {
			output, _, err := cache.ApplyPairs(word, *prefix, names...)
			return output, err
		},
		"Execute": func(templ, word string) (string, error) {
			wr := new(bytes.Buffer)
			err := t.ExecuteTemplate(wr, templ, word)
			return wr.String(), err
		},
		"Match": func(word, rule string) ([]sounds.Match, error) {
			parsed, err := sounds.ParseRule(rule)
			if err != nil {
				return nil, err
			}
			compiled, err := rl.CompileRule(parsed)
			if err != nil {
				return nil, err
			}
			return compiled.FindMatches(word), nil
		},
		"RuleList": func(data, word string) (string, error) {
			for _, l := range strings.Split(data, "\n") {
				err = rl.ParseRuleCat(l)
				if err != nil {
					return "", err
				}
			}
			output, _, err := rl.Apply(word)
			return output, err
		},
	})
	t, err = t.ParseFiles(filename)
	fatal(err)
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		word := input.Text()
		rl = sounds.NewRuleList()
		err = t.Execute(w, word)
		fatal(err)
		w.Flush()
	}
}
