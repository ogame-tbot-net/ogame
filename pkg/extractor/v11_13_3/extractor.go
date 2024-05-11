package v11_13_3

import (
	"bytes"

	"github.com/PuerkitoBio/goquery"
	v11_9_0 "github.com/alaingilbert/ogame/pkg/extractor/v11_9_0"
	"github.com/alaingilbert/ogame/pkg/ogame"
)

// Extractor ...
type Extractor struct {
	v11_9_0.Extractor
}

// NewExtractor ...
func NewExtractor() *Extractor {
	return &Extractor{}
}

// ExtractLfBonuses ...
func (e *Extractor) ExtractLfBonuses(pageHTML []byte) (ogame.LfBonuses, error) {
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	return e.ExtractLfBonusesFromDoc(doc)
}

// ExtractLfBonusesFromDoc ...
func (e *Extractor) ExtractLfBonusesFromDoc(doc *goquery.Document) (ogame.LfBonuses, error) {
	return extractLfBonusesFromDoc(doc)
}
