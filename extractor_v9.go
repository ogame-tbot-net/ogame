package ogame

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"time"
)

// ExtractorV9 ...
type ExtractorV9 struct {
	ExtractorV874
}

// NewExtractorV9 ...
func NewExtractorV9() *ExtractorV9 {
	return &ExtractorV9{}
}

// ExtractEspionageReport ...
func (e ExtractorV9) ExtractEspionageReport(pageHTML []byte, location *time.Location) (EspionageReport, error) {
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	return e.ExtractEspionageReportFromDoc(doc, location)
}

// ExtractEspionageReportFromDoc ...
func (e ExtractorV9) ExtractEspionageReportFromDoc(doc *goquery.Document, location *time.Location) (EspionageReport, error) {
	return extractEspionageReportFromDocV9(doc, location)
}

// ExtractResources ...
func (e ExtractorV9) ExtractResources(pageHTML []byte) Resources {
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	return e.ExtractResourcesFromDoc(doc)
}

// ExtractResourcesFromDoc ...
func (e ExtractorV9) ExtractResourcesFromDoc(doc *goquery.Document) Resources {
	return extractResourcesFromDocV9(doc)
}

// ExtractResourcesDetailsFromFullPage ...
func (e ExtractorV9) ExtractResourcesDetailsFromFullPage(pageHTML []byte) ResourcesDetails {
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	return e.ExtractResourcesDetailsFromFullPageFromDoc(doc)
}

// ExtractResourcesDetailsFromFullPageFromDoc ...
func (e ExtractorV9) ExtractResourcesDetailsFromFullPageFromDoc(doc *goquery.Document) ResourcesDetails {
	return extractResourcesDetailsFromFullPageFromDocV9(doc)
}
