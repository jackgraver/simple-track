package grocery

import (
	"regexp"
	"strings"
	"unicode"
)

// fuzzyThreshold is the minimum normalized edit-distance similarity (0..1)
// for a typo-level match like "bananna" -> "Banana".
const fuzzyThreshold = 0.8

// minFuzzyKeyLen avoids fuzzy matching very short keys ("egg" vs "eel").
const minFuzzyKeyLen = 4

var (
	// Leading list decorations: bullets, dashes, checkboxes, "1." / "2)" numbering.
	leadingMarkerRe = regexp.MustCompile(`^(?:[-*•·–—>+]+|\[[ xX]?\]|[☐☑✓✔]|\d+[.)])\s*`)
	splitRe         = regexp.MustCompile(`[\n\r,;]+`)
)

// Words dropped from the front of a line when building a match key, so that
// "2 lbs ground beef" and "a dozen eggs" match "Ground Beef" and "Eggs".
var quantityWords = map[string]struct{}{
	"a": {}, "an": {}, "some": {}, "of": {}, "x": {},
	"lb": {}, "lbs": {}, "pound": {}, "pounds": {}, "kg": {}, "kgs": {}, "g": {}, "oz": {},
	"l": {}, "ml": {}, "litre": {}, "litres": {}, "liter": {}, "liters": {},
	"gallon": {}, "gallons": {}, "gal": {}, "dozen": {}, "doz": {},
	"pack": {}, "packs": {}, "pk": {}, "bag": {}, "bags": {}, "box": {}, "boxes": {},
	"can": {}, "cans": {}, "bottle": {}, "bottles": {}, "jar": {}, "jars": {},
	"carton": {}, "cartons": {}, "loaf": {}, "loaves": {}, "bunch": {}, "bunches": {},
	"head": {}, "heads": {}, "container": {}, "containers": {}, "tub": {}, "tubs": {},
}

// ParseList splits free text (e.g. a pasted text message) into cleaned item
// lines. Lines are split on newlines, commas and semicolons; bullets and
// numbering are stripped; case-insensitive duplicates are dropped.
func ParseList(text string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, part := range splitRe.Split(text, -1) {
		line := normalizeName(part)
		for {
			stripped := leadingMarkerRe.ReplaceAllString(line, "")
			if stripped == line {
				break
			}
			line = stripped
		}
		line = strings.TrimRight(line, ".!?:")
		line = normalizeName(line)
		if line == "" {
			continue
		}
		key := strings.ToLower(line)
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, line)
	}
	return out
}

// matchKey reduces a name to a comparison key: lowercase, punctuation removed,
// leading quantities/units removed, each word naively singularized.
func matchKey(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		case r == '/' || r == '.' || r == '%':
			b.WriteRune(r) // keep "1/2", "1.5" intact so they're dropped as quantities
		default:
			b.WriteRune(' ')
		}
	}
	words := strings.Fields(b.String())

	// Drop leading quantity tokens ("2", "1/2", "2x", "a", "dozen", "lbs", "of").
	i := 0
	for i < len(words)-1 && isQuantityWord(words[i]) {
		i++
	}
	words = words[i:]

	for j, w := range words {
		w = strings.Trim(w, "./")
		words[j] = singularize(w)
	}
	return strings.Join(strings.Fields(strings.Join(words, " ")), " ")
}

func isQuantityWord(w string) bool {
	if _, ok := quantityWords[w]; ok {
		return true
	}
	// Numbers, fractions, decimals, "2x", "x2", "500g", "2lb".
	digits := 0
	for _, r := range w {
		if unicode.IsDigit(r) {
			digits++
		}
	}
	if digits == 0 {
		return false
	}
	rest := strings.TrimLeft(w, "0123456789./")
	rest = strings.TrimPrefix(rest, "x")
	if rest == "" {
		return true
	}
	if strings.HasPrefix(w, "x") && strings.Trim(w[1:], "0123456789") == "" {
		return true
	}
	_, ok := quantityWords[rest]
	return ok
}

func singularize(w string) string {
	n := len(w)
	switch {
	case n > 4 && strings.HasSuffix(w, "ies"):
		return w[:n-3] + "y"
	case n > 4 && strings.HasSuffix(w, "oes"):
		return w[:n-2]
	case n > 4 && (strings.HasSuffix(w, "ches") || strings.HasSuffix(w, "shes") ||
		strings.HasSuffix(w, "xes") || strings.HasSuffix(w, "sses")):
		return w[:n-2]
	case n > 3 && strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss") &&
		!strings.HasSuffix(w, "us") && !strings.HasSuffix(w, "is"):
		return w[:n-1]
	}
	return w
}

// similarity returns 1 - levenshtein(a,b)/max(len(a),len(b)).
func similarity(a, b string) float64 {
	ra, rb := []rune(a), []rune(b)
	if len(ra) == 0 && len(rb) == 0 {
		return 1
	}
	prev := make([]int, len(rb)+1)
	cur := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return 1 - float64(prev[len(rb)])/float64(max(len(ra), len(rb)))
}

// KnownItem is a previously used grocery name and how often it was added.
type KnownItem struct {
	Name  string
	Count int
}

// ParsedLine is one parsed line from a pasted list with its suggested name.
type ParsedLine struct {
	Input   string `json:"input"`
	Name    string `json:"name"`
	Matched bool   `json:"matched"`
	OnList  bool   `json:"on_list"`
}

// MatchLines resolves each parsed line against known item history and the
// currently active list. known should be ordered most-preferred first.
func MatchLines(lines []string, known []KnownItem, active []string) []ParsedLine {
	type candidate struct {
		name  string
		key   string
		count int
	}
	byKey := make(map[string]candidate)
	cands := make([]candidate, 0, len(known))
	for _, k := range known {
		key := matchKey(k.Name)
		if key == "" {
			continue
		}
		if existing, ok := byKey[key]; ok && existing.count >= k.Count {
			continue
		}
		c := candidate{name: k.Name, key: key, count: k.Count}
		byKey[key] = c
	}
	for _, c := range byKey {
		cands = append(cands, c)
	}

	activeKeys := make(map[string]struct{}, len(active))
	for _, a := range active {
		activeKeys[matchKey(a)] = struct{}{}
	}

	out := make([]ParsedLine, 0, len(lines))
	for _, line := range lines {
		key := matchKey(line)
		res := ParsedLine{Input: line, Name: line}

		if c, ok := byKey[key]; ok {
			res.Name, res.Matched = c.name, true
		} else if len([]rune(key)) >= minFuzzyKeyLen {
			var best *candidate
			bestScore := 0.0
			for i := range cands {
				c := &cands[i]
				if len([]rune(c.key)) < minFuzzyKeyLen {
					continue
				}
				s := similarity(key, c.key)
				if s < fuzzyThreshold {
					continue
				}
				if best == nil || s > bestScore || (s == bestScore && c.count > best.count) ||
					(s == bestScore && c.count == best.count && c.name < best.name) {
					best, bestScore = c, s
				}
			}
			if best != nil {
				res.Name, res.Matched = best.name, true
			}
		}

		if _, ok := activeKeys[matchKey(res.Name)]; ok {
			res.OnList = true
		}
		out = append(out, res)
	}
	return out
}
