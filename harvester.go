package harvester

import (
	"net/url"
	"regexp"
)

type IgnoreDiscoveredResourceRule interface {
	IgnoreDiscoveredResource(url *url.URL) (bool, string)
}

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
func MakeContentHarvester(discoverURLsRegEx *regexp.Regexp, ignoreResourceRule IgnoreDiscoveredResourceRule, cleanResourceRule CleanDiscoveredResourceRule) *ContentHarvester {
	result := new(ContentHarvester)
	result.discoverURLsRegEx = discoverURLsRegEx
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
