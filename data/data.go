// Package data embeds the Bible texts and exposes them as loaded translations.
package data

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/KalleNgithub/bible-bot/internal/bible"
)

//go:embed web.tsv
var webTSV []byte

//go:embed fi.tsv
var fiTSV []byte

//go:embed vectors.bin
var vectorsBin []byte

// Vectors returns the embedded verse vectors file.
func Vectors() *bytes.Reader { return bytes.NewReader(vectorsBin) }

// Load returns the English (WEB) and Finnish (1933/38) translations.
func Load() (en, fi *bible.Translation, err error) {
	if en, err = bible.Load(bytes.NewReader(webTSV)); err != nil {
		return nil, nil, fmt.Errorf("web: %w", err)
	}
	if fi, err = bible.Load(bytes.NewReader(fiTSV)); err != nil {
		return nil, nil, fmt.Errorf("fi: %w", err)
	}
	return en, fi, nil
}
