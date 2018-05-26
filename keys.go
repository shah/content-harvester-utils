package harvester

import (
	"github.com/Machiel/slugify"
	"github.com/julianshen/og"
)

// HarvestedResourceKeys tracks the keys for a single URL that was discovered in content.
// Keys allow the URL to be identified in a database, key value store, etc.
type HarvestedResourceKeys struct {
	hr       *HarvestedResource
	pageInfo *og.PageInfo
	piError  error
}

// IsValid returns true if there are no errors
func (hrk *HarvestedResourceKeys) IsValid() bool {
	return hrk.piError == nil
}

// Slug returns the title of the content
func (hrk *HarvestedResourceKeys) Slug() string {
	if hrk.piError == nil {
		return slugify.Slugify(hrk.pageInfo.Title)
	}
	return "Error getting PageInfo"
}

// CreateKeys returns a new resource keys object
func CreateKeys(hr *HarvestedResource) *HarvestedResourceKeys {
	result := new(HarvestedResourceKeys)
	result.hr = hr
	result.pageInfo, result.piError = og.GetPageInfoFromUrl(hr.finalURL.String())
	return result
}
