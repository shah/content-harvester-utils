package harvester

import (
	"net/url"
	"regexp"

	"mvdan.cc/xurls"
)

// TODO use https://github.com/PuerkitoBio/goquery for parsing singe page HTML (similar to cheerio library for Node.js)
// TODO use https://github.com/gocolly/colly for scraping multiple page HTML sites
// TODO use https://github.com/andrewstuart/goq for type-safe layer on top of goquery using struct-tag
// TODO use https://github.com/spf13/afero as a fileSystem abstraction layer in case resources will be stored
// TODO use https://github.com/h2non/filetype to infer file types checking the magic numbers signature (after downloading the file and keeping it the filesystem)
// TODO use https://medium.com/@aschers/deploy-machine-learning-models-from-r-research-to-ruby-go-production-with-pmml-b41e79445d3d for scoring content

// IgnoreDiscoveredResourceRule is a rule
type IgnoreDiscoveredResourceRule interface {
	IgnoreDiscoveredResource(url *url.URL) (bool, string)
}

// CleanDiscoveredResourceRule is a rule
type CleanDiscoveredResourceRule interface {
	CleanDiscoveredResource(url *url.URL) bool
	RemoveQueryParamFromResource(paramName string) (bool, string)
}

// ContentHarvester discovers URLs (called "Resources" from the "R" in "URL")
type ContentHarvester struct {
	discoverURLsRegEx  *regexp.Regexp
	ignoreResourceRule IgnoreDiscoveredResourceRule
	cleanResourceRule  CleanDiscoveredResourceRule
}

// The HarvestedResources is the list of URLs discovered in a piece of content
type HarvestedResources struct {
	// TODO remove duplicates in case the same resource is included more than once
	Resources []*HarvestedResource
}

// MakeContentHarvester prepares a content harvester
func MakeContentHarvester(ignoreResourceRule IgnoreDiscoveredResourceRule, cleanResourceRule CleanDiscoveredResourceRule) *ContentHarvester {
	result := new(ContentHarvester)
	result.discoverURLsRegEx = xurls.Relaxed
	result.ignoreResourceRule = ignoreResourceRule
	result.cleanResourceRule = cleanResourceRule
	return result
}

// HarvestResources discovers URLs within content and returns what was found
func (h *ContentHarvester) HarvestResources(content string) *HarvestedResources {
	result := new(HarvestedResources)
	urls := h.discoverURLsRegEx.FindAllString(content, -1)
	for _, urlText := range urls {
		res := harvestResource(h, urlText)

		// TODO check for duplicates and only append unique discovered URLs
		result.Resources = append(result.Resources, res)
	}
	return result
}
