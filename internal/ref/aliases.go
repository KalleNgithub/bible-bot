package ref

// Hand-written aliases that prefix matching on the canonical names does not cover.
// Numbered books use a leading digit, e.g. "1 jn".
var englishExtras = map[int][]string{
	2:  {"ex"},
	19: {"ps", "psalm", "pslm"},
	22: {"song", "song of solomon", "sos", "canticles"},
	31: {"obad"},
	32: {"jon"},
	40: {"mt"},
	41: {"mk"},
	42: {"lk"},
	43: {"jn"},
	44: {"acts of the apostles"},
	50: {"phil", "php"},
	57: {"phm"},
	58: {"heb"},
	62: {"1 jn"},
	63: {"2 jn"},
	64: {"3 jn"},
	65: {"jude", "jd"},
	66: {"revelations", "rev", "apocalypse"},
}

var finnishExtras = map[int][]string{
	19: {"ps"},
	21: {"saarn"},
	31: {"ob"},
	40: {"matteus"},
	41: {"markus"},
	42: {"luukas"},
	43: {"johannes"},
	44: {"apt"},
	48: {"gal"},
	49: {"ef"},
	50: {"fil"},
	57: {"filem"},
	58: {"hepr"},
	66: {"ilm"},
}

var ordinalWords = map[string]int{
	"first": 1, "1st": 1, "i": 1, "ensimmainen": 1,
	"second": 2, "2nd": 2, "ii": 2, "toinen": 2,
	"third": 3, "3rd": 3, "iii": 3, "kolmas": 3,
	"neljas": 4, "viides": 5,
}

// Words that carry no book identity.
var fillerWords = map[string]bool{
	"chapter": true, "chap": true, "ch": true, "luku": true, "kappale": true,
	"verse": true, "verses": true, "v": true, "vv": true, "jae": true, "jakeet": true,
}

// Words joining two verse numbers into a range; "and" only counts for adjacent verses.
var rangeWords = map[string]bool{
	"to": true, "through": true, "thru": true, "till": true, "until": true,
	"asti": true, "saakka": true, "and": true, "ja": true,
}

var nameSuffixWords = map[string]bool{"kirja": true, "evankeliumi": true, "kirje": true}
