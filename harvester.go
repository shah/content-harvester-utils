package harvester

import (
	"regexp"

	"mvdan.cc/xurls"
)

type ContentHarvester struct {
	discoverURLsRegEx         *regexp.Regexp
	ignoreURLsRegEx           []*regexp.Regexp
	removeParamsFromURLsRegEx []*regexp.Regexp
}

func MakeDefaultContentHarvester() *ContentHarvester {
	// setup matchers -- use https://regex-golang.appspot.com/assets/html/index.html for testing
	ignoreURLsRegEx := []*regexp.Regexp{regexp.MustCompile(`^https://twitter.com/(.*?)/status/(.*)$`), regexp.MustCompile(`https://t.co`)}
	removeParamsFromURLsRegEx := []*regexp.Regexp{regexp.MustCompile(`^utm_`)}
	return MakeContentHarvester(xurls.Relaxed(), ignoreURLsRegEx, removeParamsFromURLsRegEx)
}

func MakeContentHarvester(discoverURLsRegEx *regexp.Regexp, ignoreURLsRegEx []*regexp.Regexp, removeParamsFromURLsRegEx []*regexp.Regexp) *ContentHarvester {
	result := new(ContentHarvester)
	result.discoverURLsRegEx = discoverURLsRegEx
	result.ignoreURLsRegEx = ignoreURLsRegEx
	result.removeParamsFromURLsRegEx = removeParamsFromURLsRegEx
	return result
}

type HarvestedResources struct {
	Resources []*Resource
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
