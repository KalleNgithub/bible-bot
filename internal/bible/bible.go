package bible

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// Ref identifies one verse. Book is 1-66 (Genesis=1, John=43).
type Ref struct {
	Book, Chapter, Verse int
}

func (r Ref) key() int { return r.Book*1000000 + r.Chapter*1000 + r.Verse }

// RefFromKey decodes a verse key as stored in the vector file.
func RefFromKey(k uint32) Ref {
	return Ref{Book: int(k / 1000000), Chapter: int(k / 1000 % 1000), Verse: int(k % 1000)}
}

// Translation holds one Bible version in memory.
type Translation struct {
	Books  [67]string
	verses map[int]string
	refs   []Ref
	maxV   map[int]int
}

func New() *Translation {
	return &Translation{verses: map[int]string{}, maxV: map[int]int{}}
}

func (t *Translation) Add(r Ref, text string) {
	t.verses[r.key()] = text
	t.refs = append(t.refs, r)
	ck := r.Book*1000 + r.Chapter
	if r.Verse > t.maxV[ck] {
		t.maxV[ck] = r.Verse
	}
}

func (t *Translation) SetBook(nr int, name string) { t.Books[nr] = name }

// Text returns the verse text and whether the verse exists.
func (t *Translation) Text(r Ref) (string, bool) {
	s, ok := t.verses[r.key()]
	return s, ok
}

// LastVerse returns the highest verse number in a chapter, or 0 if the chapter does not exist.
func (t *Translation) LastVerse(book, chapter int) int { return t.maxV[book*1000+chapter] }

// Refs returns all verse references in canonical order.
func (t *Translation) Refs() []Ref {
	sort.Slice(t.refs, func(i, j int) bool { return t.refs[i].key() < t.refs[j].key() })
	return t.refs
}

// Write serialises as tab-separated lines: "B<TAB>nr<TAB>name" for books,
// then "V<TAB>book<TAB>chapter<TAB>verse<TAB>text" for verses.
func (t *Translation) Write(w io.Writer) error {
	bw := bufio.NewWriter(w)
	for nr := 1; nr <= 66; nr++ {
		if t.Books[nr] != "" {
			fmt.Fprintf(bw, "B\t%d\t%s\n", nr, t.Books[nr])
		}
	}
	for _, r := range t.Refs() {
		fmt.Fprintf(bw, "V\t%d\t%d\t%d\t%s\n", r.Book, r.Chapter, r.Verse, t.verses[r.key()])
	}
	return bw.Flush()
}

// Load reads the format produced by Write.
func Load(r io.Reader) (*Translation, error) {
	t := New()
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	line := 0
	for sc.Scan() {
		line++
		f := strings.SplitN(sc.Text(), "\t", 5)
		switch f[0] {
		case "B":
			if len(f) != 3 {
				return nil, fmt.Errorf("line %d: bad book line", line)
			}
			nr, err := strconv.Atoi(f[1])
			if err != nil || nr < 1 || nr > 66 {
				return nil, fmt.Errorf("line %d: bad book number", line)
			}
			t.SetBook(nr, f[2])
		case "V":
			if len(f) != 5 {
				return nil, fmt.Errorf("line %d: bad verse line", line)
			}
			var n [3]int
			for i := 0; i < 3; i++ {
				v, err := strconv.Atoi(f[i+1])
				if err != nil {
					return nil, fmt.Errorf("line %d: bad number %q", line, f[i+1])
				}
				n[i] = v
			}
			t.Add(Ref{n[0], n[1], n[2]}, f[4])
		case "":
		default:
			return nil, fmt.Errorf("line %d: unknown record %q", line, f[0])
		}
	}
	return t, sc.Err()
}
