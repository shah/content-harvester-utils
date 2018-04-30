package harvester

import (
	"fmt"
	"net/url"
	"regexp"
)

// ContentHarvester discovers URLs (called "Resources" from the "R" in "URL")
type ContentHarvester struct {
	discoverURLsRegEx         *regexp.Regexp
	ignoreURLsRegEx           []*regexp.Regexp
	removeParamsFromURLsRegEx []*regexp.Regexp
}

// The HarvestedResources is the list of URLs discovered in a piece of content
type HarvestedResources struct {
	// TODO remove duplicates in case the same resource is included more than once
	Resources []*HarvestedResource
}

// MakeContentHarvester prepares a content harvester
func MakeContentHarvester(discoverURLsRegEx *regexp.Regexp, ignoreURLsRegEx []*regexp.Regexp, removeParamsFromURLsRegEx []*regexp.Regexp) *ContentHarvester {
	result := new(ContentHarvester)
	result.discoverURLsRegEx = discoverURLsRegEx
	result.ignoreURLsRegEx = ignoreURLsRegEx
	result.removeParamsFromURLsRegEx = removeParamsFromURLsRegEx
	return result
}

func (h *ContentHarvester) ignoreResource(url *url.URL) (bool, string) {
	URLtext := url.String()
	for _, regEx := range h.ignoreURLsRegEx {
		if regEx.MatchString(URLtext) {
			return true, fmt.Sprintf("Matched Ignore Rule `%s`", regEx.String())
		}
	}

	return false, ""
}

func (h *ContentHarvester) cleanResource(url *url.URL) (bool, *url.URL) {
	harvestedParams := url.Query()
	type ParamMatch struct {
		paramName string
		regEx     *regexp.Regexp
	}
	var cleanedParams []ParamMatch
	for paramName := range harvestedParams {
		for _, regEx := range h.removeParamsFromURLsRegEx {
			if regEx.MatchString(paramName) {
				harvestedParams.Del(paramName)
				cleanedParams = append(cleanedParams, ParamMatch{paramName, regEx})
			}
		}
	}

	if len(cleanedParams) > 0 {
		cleanedURL := url
		cleanedURL.RawQuery = harvestedParams.Encode()
		return true, cleanedURL
	}
	return false, url
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
