package main

import (
	"bufio"
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

	filename := flag.String("i", "", "template file")
	flag.Parse()

	rl := sounds.NewRuleList()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	t := template.New(filepath.Base(*filename))
	t = t.Funcs(template.FuncMap{
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
	t, err = t.ParseFiles(*filename)
	fatal(err)
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		word := input.Text()
		err = t.Execute(w, word)
		fatal(err)
		w.Flush()
		rl = sounds.NewRuleList()
	}
}
