package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/zyxw59/conlang/sounds"
	"log"
	"os"
	"strings"
)

func main() {
	verbose := flag.Bool("v", false, "verbose: print debug output")
	quiet := flag.Bool("q", false, "quiet: do not print prompts")
	prefix := flag.String("p", "", "prefix for sound change files")

	flag.Parse()

	pairs := flag.Args()
	cache := sounds.NewCache()
	_, err := cache.LoadPairs(*prefix, pairs...)
	if err != nil {
		log.Fatal(err)
	}
	if !*quiet {
		fmt.Println("Type words to apply changes to. ^C to quit")
	}
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		word := input.Text()
		output, debug, err := cache.ApplyPairs(word, *prefix, pairs...)
		if err != nil {
			log.Fatal(err)
		}
		if *verbose {
			fmt.Println(strings.Join(debug, "\n"))
		}
		fmt.Println(output)
	}
}
