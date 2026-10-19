// Package ref parses loosely written Bible references such as "Jhon 11:25 to 26".
package ref

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/KalleNgithub/bible-bot/internal/bible"
)

// ErrNotReference means the text does not look like a verse reference at all.
var ErrNotReference = errors.New("not a verse reference")

// NotFoundError means the text looked like a reference but names no existing verse.
type NotFoundError struct{ Msg string }

func (e *NotFoundError) Error() string { return e.Msg }

// Result is a resolved verse range within one chapter.
type Result struct {
	Book, Chapter, From, To int
	Finnish                 bool // the book was named with a Finnish-only alias
}

type alias struct {
	ord  int
	name string
	book int
	fi   bool
}

type Parser struct {
	aliases []alias
	en      *bible.Translation
}

// NewParser builds the alias table from the book names of both translations.
func NewParser(en, fi *bible.Translation) *Parser {
	p := &Parser{en: en}
	for nr := 1; nr <= 66; nr++ {
		p.addName(en.Books[nr], nr, false)
		p.addName(fi.Books[nr], nr, true)
		for _, a := range englishExtras[nr] {
			p.addName(a, nr, false)
		}
		for _, a := range finnishExtras[nr] {
			p.addName(a, nr, true)
		}
	}
	return p
}

func (p *Parser) addName(name string, book int, fi bool) {
	toks := tokenize(name)
	ord := 0
	if len(toks) > 1 {
		if o, ok := ordinalOf(toks[0]); ok {
			ord, toks = o, toks[1:]
		}
	}
	var words []string
	for _, t := range toks {
		if isWord(t) {
			words = append(words, t)
		}
	}
	if len(words) == 0 {
		return
	}
	p.aliases = append(p.aliases, alias{ord, strings.Join(words, ""), book, fi})
	if len(words) > 1 && nameSuffixWords[words[len(words)-1]] {
		p.aliases = append(p.aliases, alias{ord, strings.Join(words[:len(words)-1], ""), book, fi})
	}
}

func ordinalOf(tok string) (int, bool) {
	if n, ok := ordinalWords[tok]; ok {
		return n, true
	}
	if n, err := strconv.Atoi(tok); err == nil && n >= 1 && n <= 5 {
		return n, true
	}
	return 0, false
}

func isWord(t string) bool { return t[0] >= 'a' && t[0] <= 'z' }
func isNum(t string) bool  { return t[0] >= '0' && t[0] <= '9' }
func isDash(t string) bool { return t == "-" }

var fold = strings.NewReplacer("ä", "a", "å", "a", "ö", "o", "é", "e", "–", "-", "—", "-")

// tokenize lowercases, folds Nordic letters and splits into words, numbers and dashes.
// Everything else (colons, dots, commas) is dropped.
func tokenize(s string) []string {
	s = fold.Replace(strings.ToLower(s))
	var toks []string
	var cur []byte
	kind := byte(0)
	flush := func() {
		if len(cur) > 0 {
			toks = append(toks, string(cur))
			cur = cur[:0]
		}
		kind = 0
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		var k byte
		switch {
		case c >= 'a' && c <= 'z':
			k = 'a'
		case c >= '0' && c <= '9':
			k = '0'
		case c == '-':
			flush()
			toks = append(toks, "-")
			continue
		default:
			flush()
			continue
		}
		if k != kind {
			flush()
			kind = k
		}
		cur = append(cur, c)
	}
	flush()
	return toks
}

// Parse resolves a reference. It returns ErrNotReference when the text is not
// shaped like one, and *NotFoundError when it is but the verse does not exist.
func (p *Parser) Parse(q string) (Result, error) {
	toks := tokenize(q)
	i, ord := 0, 0
	if len(toks) > 1 && isWord(toks[1]) {
		if o, ok := ordinalOf(toks[0]); ok {
			ord, i = o, 1
		}
	}
	var words []string
	for ; i < len(toks) && isWord(toks[i]); i++ {
		words = append(words, toks[i])
	}
	for len(words) > 0 && fillerWords[words[len(words)-1]] {
		words = words[:len(words)-1]
	}
	if len(words) == 0 || len(words) > 4 {
		return Result{}, ErrNotReference
	}

	var nums []int
	var gaps []string // gaps[k] joins nums[k] and nums[k+1]: "", "-" or a range word
	pending := ""
	for ; i < len(toks); i++ {
		t := toks[i]
		switch {
		case isNum(t):
			n, err := strconv.Atoi(t)
			if err != nil {
				return Result{}, ErrNotReference
			}
			if len(nums) > 0 {
				gaps = append(gaps, pending)
			}
			pending = ""
			nums = append(nums, n)
		case isDash(t):
			pending = "-"
		case rangeWords[t]:
			if pending == "" {
				pending = t
			}
		case fillerWords[t]:
		default:
			return Result{}, ErrNotReference
		}
	}
	if len(nums) == 0 {
		return Result{}, ErrNotReference
	}

	book, fi, ok := p.matchBook(ord, strings.Join(words, ""))
	if !ok {
		return Result{}, ErrNotReference
	}
	name := p.en.Books[book]

	singleChapter := p.en.LastVerse(book, 2) == 0
	r := Result{Book: book, Finnish: fi}
	switch {
	case len(nums) == 1 && singleChapter:
		r.Chapter, r.From, r.To = 1, nums[0], nums[0]
	case len(nums) == 1:
		return Result{}, &NotFoundError{fmt.Sprintf("Give a verse as well, e.g. %s %d:1", name, nums[0])}
	case len(nums) == 2 && singleChapter && gaps[0] != "":
		r.Chapter, r.From, r.To = 1, nums[0], nums[1]
	case len(nums) == 2:
		r.Chapter, r.From, r.To = nums[0], nums[1], nums[1]
	case len(nums) == 3:
		r.Chapter, r.From, r.To = nums[0], nums[1], nums[2]
		if gaps[1] == "and" || gaps[1] == "ja" {
			if r.To != r.From+1 {
				return Result{}, ErrNotReference
			}
		}
	default:
		return Result{}, ErrNotReference
	}

	last := p.en.LastVerse(book, r.Chapter)
	switch {
	case last == 0:
		return Result{}, &NotFoundError{fmt.Sprintf("%s has no chapter %d", name, r.Chapter)}
	case r.From < 1 || r.From > last || r.To > last:
		return Result{}, &NotFoundError{fmt.Sprintf("%s %d has no such verse (last is %d)", name, r.Chapter, last)}
	case r.To < r.From:
		return Result{}, &NotFoundError{fmt.Sprintf("Verse range %d-%d is backwards", r.From, r.To)}
	}
	return r, nil
}

// matchBook resolves a book from an optional ordinal and the compacted name words:
// exact alias first, then unique prefix, then unique nearest by edit distance.
func (p *Parser) matchBook(ord int, name string) (book int, fi, ok bool) {
	if exact := p.collect(func(a alias) bool { return a.ord == ord && a.name == name }); len(exact) > 0 {
		return resolve(exact)
	}
	if len(name) < 3 {
		return 0, false, false
	}
	if pre := p.collect(func(a alias) bool { return a.ord == ord && strings.HasPrefix(a.name, name) }); len(pre) > 0 {
		return resolve(pre)
	}

	limit := 2
	if len(name) <= 4 {
		limit = 1
	}
	best := limit + 1
	var cands []alias
	for _, a := range p.aliases {
		if a.ord != ord {
			continue
		}
		d := distance(name, a.name)
		if d > limit || d > best {
			continue
		}
		if d < best {
			best, cands = d, nil
		}
		cands = append(cands, a)
	}
	if len(cands) == 0 {
		return 0, false, false
	}
	if sub := filter(cands, func(a alias) bool { return isSubsequence(name, a.name) || sameLetters(name, a.name) }); len(sub) > 0 {
		cands = sub
	}
	return resolve(cands)
}

func (p *Parser) collect(f func(alias) bool) []alias { return filter(p.aliases, f) }

func filter(in []alias, f func(alias) bool) []alias {
	var out []alias
	for _, a := range in {
		if f(a) {
			out = append(out, a)
		}
	}
	return out
}

// resolve succeeds only when all candidates name the same book.
func resolve(c []alias) (int, bool, bool) {
	fi := true
	for _, a := range c {
		if a.book != c[0].book {
			return 0, false, false
		}
		fi = fi && a.fi
	}
	return c[0].book, fi, true
}

// sameLetters reports whether s is a permutation of t (covers transposition typos).
func sameLetters(s, t string) bool {
	if len(s) != len(t) {
		return false
	}
	var n [256]int
	for i := 0; i < len(s); i++ {
		n[s[i]]++
		n[t[i]]--
	}
	return n == [256]int{}
}

func isSubsequence(s, t string) bool {
	j := 0
	for i := 0; i < len(t) && j < len(s); i++ {
		if t[i] == s[j] {
			j++
		}
	}
	return j == len(s)
}

// distance is the optimal string alignment distance (insert, delete, substitute, adjacent swap).
func distance(a, b string) int {
	la, lb := len(a), len(b)
	d := make([][]int, la+1)
	for i := range d {
		d[i] = make([]int, lb+1)
		d[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		d[0][j] = j
	}
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			d[i][j] = min(d[i-1][j]+1, d[i][j-1]+1, d[i-1][j-1]+cost)
			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				d[i][j] = min(d[i][j], d[i-2][j-2]+1)
			}
		}
	}
	return d[la][lb]
}
