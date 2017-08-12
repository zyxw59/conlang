# Conlang: Language construction utilities

This project is a collection of tools for working on constructed languages
(currently, there is only one tool, but in the future, I plan to include more).

## Soundchanger

The `sounds` package and the `soundchanger` program provide tools to apply
lists of sound changes to words. This tool can be used in diachronic conlanging
to systematically evolve words from one language to another, or to apply
synchronic changes.

### How to use

#### Sound change files

Sound change files used by this program are read line-by-line. Each line can be
one of the following:

##### A sound change rule
A sound change rule has three parts:
- The change, written as _a_` > `_b_, meaning that _a_ will become _b_
- The environment, written as ` / `_c_`_`_d_, meaning that the change will only
  occur when _a_ occurs between _c_ and _d_
- The negative environment, written as ` ! `_e_`_`_f_, meaning that the change
  will not occur if _a_ occurs after _e_ or before _f_

Only the change is required, the other two sections are optional.

For the original sound, the environment, and the negative environment
(components _a_, _c_, _d_, _e_, and _f_ of the rule), [Regular Expression
syntax](https://github.com/google/re2/wiki/Syntax) can be used. This program
includes some additional features as well:
- Categories: Categories, which represent a set of sounds, can be included
  using the syntax `{`_categoryName_`}`, which will match any element of the
  category. Categories must be defined before they are used, as described
  [below](#a-category-definition)
- Numbered categories: Categories can also be numbered, which forces all
  categories with the same number to match the same element index. For example,
  if `N` is the category `m n ŋ`, and `P` is the category `p t k`, `{0:N}{0:P}`
  will match `mp`, `nt`, and `ŋk`, but not things like `mt` or `ŋp` (those
          would still be matched by `{N}{P}`)
  - If a numbered category is included in the result of the sound change
    (component _b_), it will be replaced by the appropriate value of that
    category. For example (continuing from above), the rule `{0:P} > {0:N}`
    will cause `p` to become `m`, `t` to become `n`, and `k` to become `ŋ`.
- Word boundaries: The standard Regex `\b` only correctly matches ASCII word
  boundaries, which is generally not sufficient for conlinguists who make
  heavy use of Unicode. Instead, this program offers the character `#`, which
  can be used to match word boundaries in the environment or negative
  environment of a rule. It matches only a boundary between whitespace and
  non-whitespace.

##### A category definition
A category definition has the following format: _name_` = `_elements_, where
_name_ is the name of the category, and _elements_ is a whitespace-separated
list of the elements of that category. A category can also include another
(previously-defined) category as an element, in which case that category is
expanded into its elements, which are then included.

##### A comment
A comment is a line that starts with `//`. It has no effect on the running of
the program, but will be output with the debugging info to provide context.
Writing comments is highly recommended, as it makes it easier to tell at a
glance what each rule is supposed to do. (Source: personal experience)

##### A blank line
Blank lines are ignored.

#### `soundchanger`

The `soundchanger` program is a simple tool to apply one or more sound change
files to words.

##### Basic usage
```
soundchanger [-v] [-q] [-p _prefix_] _pairs_
```
- `-v` verbose mode: output debug info as along with the words
- `-q` quiet mode: don't print initial prompt
- `-p` _prefix_: use _prefix_ as a prefix before all filenames
- _pairs_: a list of whitespace-separated pairs of languages, as described
  [below](#file-structure)

Once `soundchanger` is running, it reads lines from `stdin`, applies changes,
and outputs on `stdout`. Note that if you update any of the sound change files
while `soundchanger` is running, it will automatically re-read the file, so you
don't need to restart the program in this case.

##### File structure
To describe language trees, `soundchanger` uses dot-separated file names for
sound changes. For example, a set of files for describing the changes from
Latin to the Romance Languages might something like this (this is of course
heavily simplified for use as an example):
```
latin
latin.ecclesiastical
latin.vulgar
latin.vulgar
latin.vulgar.french
latin.vulgar.french.creyol
latin.vulgar.iberian
latin.vulgar.iberian.portuguese
latin.vulgar.iberian.spanish
latin.vulgar.italian
latin.vulgar.romanian
```
Then, to apply changes to arrive at Spanish from Latin, you would apply the
files `latin.vulgar`, `latin.vulgar.iberian`, and
`latin.vulgar.iberian.spanish` in order. To simplify this task, you can instead
specify the start and end point of this chain, in this case `latin` and
`latin.vulgar.iberian.spanish`, and `soundchanger` will automatically apply all
the files in the correct sequence. So the full invocation of `soundchanger`
would be:
```
soundchanger latin latin.vulgar.iberian.spanish
```

