package harvester

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
)

type RemovedResourceParam struct {
	paramName  string
	matchRegEx *regexp.Regexp
}

type Resource struct {
	origURLtext      string
	isURLValid       bool
	isDestValid      bool
	isURLIgnored     bool
	ignoreReason     string
	isURLCleaned     bool
	urlParamsCleaned []RemovedResourceParam
	destContentType  string
	resolvedURL      *url.URL
	cleanedURL       *url.URL
	finalURL         *url.URL
}

func (r *Resource) OriginalURLText() string {
	return r.origURLtext
}

func (r *Resource) IsValid() (bool, bool) {
	return r.isURLValid, r.isDestValid
}

func (r *Resource) IsIgnored() (bool, string) {
	return r.isURLIgnored, r.ignoreReason
}

func (r *Resource) IsCleaned() (bool, []RemovedResourceParam) {
	return r.isURLCleaned, r.urlParamsCleaned
}

func (r *Resource) GetURLs() (*url.URL, *url.URL, *url.URL) {
	return r.finalURL, r.resolvedURL, r.cleanedURL
}

func (r *Resource) DestinationContentType() string {
	return r.destContentType
}

func harvestResource(h *ContentHarvester, origURLtext string) *Resource {
	result := new(Resource)
	result.origURLtext = origURLtext

	resp, err := http.Get(origURLtext)
	result.isURLValid = err == nil
	if result.isURLValid == false {
		result.isDestValid = false
		result.isURLIgnored = true
		result.ignoreReason = fmt.Sprintf("Invalid URL '%s'", origURLtext)
		return result
	}

	if resp.StatusCode != 200 {
		result.isDestValid = false
		result.isURLIgnored = true
		result.ignoreReason = fmt.Sprintf("Invalid HTTP Status Code %d", resp.StatusCode)
		return result
	}

	result.destContentType = resp.Header.Get("Content-type")
	result.resolvedURL = resp.Request.URL
	resolvedURLText := result.resolvedURL.String()
	for _, regEx := range h.ignoreURLsRegEx {
		if regEx.MatchString(resolvedURLText) {
			result.isDestValid = true
			result.isURLIgnored = true
			result.ignoreReason = fmt.Sprintf("Matched Ignore Rule `%s`", regEx.String())
			return result
		}
	}

	result.isURLIgnored = false
	result.isDestValid = true
	cleanedParams := result.resolvedURL.Query()
	for paramName := range cleanedParams {
		for _, regEx := range h.removeParamsFromURLsRegEx {
			if regEx.MatchString(paramName) {
				cleanedParams.Del(paramName)
				result.urlParamsCleaned = append(result.urlParamsCleaned, RemovedResourceParam{paramName, regEx})
			}
		}
	}
	if len(result.urlParamsCleaned) > 0 {
		result.cleanedURL = result.resolvedURL
		result.cleanedURL.RawQuery = cleanedParams.Encode()
		result.finalURL = result.cleanedURL
		result.isURLCleaned = true
	} else {
		result.isURLCleaned = false
		result.finalURL = result.resolvedURL
	}

	// TODO once the URL is cleaned, double-check the cleaned URL to see if it's a valid destination; if not, revert to non-cleaned version
	// this could be done recursively here or by the outer function

	return result
}
