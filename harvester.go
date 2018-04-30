package harvester

import (
	"fmt"
	"net/url"
	"regexp"
)

type ContentHarvester struct {
	discoverURLsRegEx         *regexp.Regexp
	ignoreURLsRegEx           []*regexp.Regexp
	removeParamsFromURLsRegEx []*regexp.Regexp
}

func MakeContentHarvester(discoverURLsRegEx *regexp.Regexp, ignoreURLsRegEx []*regexp.Regexp, removeParamsFromURLsRegEx []*regexp.Regexp) *ContentHarvester {
	result := new(ContentHarvester)
	result.discoverURLsRegEx = discoverURLsRegEx
	result.ignoreURLsRegEx = ignoreURLsRegEx
	result.removeParamsFromURLsRegEx = removeParamsFromURLsRegEx
	return result
}

type HarvestedResources struct {
	// TODO remove duplicates in case the same resource is included more than once
	Resources []*HarvestedResource
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
	} else {
		return false, url
	}
}

func (h *ContentHarvester) HarvestResources(content string) *HarvestedResources {
	result := new(HarvestedResources)
	urls := h.discoverURLsRegEx.FindAllString(content, -1)
	for _, urlText := range urls {
		res := harvestResource(h, urlText)
		result.Resources = append(result.Resources, res)
	}
	return result
}
