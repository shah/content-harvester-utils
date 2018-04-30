package harvester

import (
	"fmt"
	"net/http"
	"net/url"
)

type HarvestedResource struct {
	// TODO consider adding source information (e.g. tweet, e-mail, etc.)
	origURLtext     string
	isURLValid      bool
	isDestValid     bool
	isURLIgnored    bool
	ignoreReason    string
	isURLCleaned    bool
	destContentType string
	resolvedURL     *url.URL
	cleanedURL      *url.URL
	finalURL        *url.URL
}

func (r *HarvestedResource) OriginalURLText() string {
	return r.origURLtext
}

func (r *HarvestedResource) IsValid() (bool, bool) {
	return r.isURLValid, r.isDestValid
}

func (r *HarvestedResource) IsIgnored() (bool, string) {
	return r.isURLIgnored, r.ignoreReason
}

func (r *HarvestedResource) IsCleaned() (bool, *url.URL) {
	return r.isURLCleaned, r.cleanedURL
}

func (r *HarvestedResource) GetURLs() (*url.URL, *url.URL, *url.URL) {
	return r.finalURL, r.resolvedURL, r.cleanedURL
}

func (r *HarvestedResource) DestinationContentType() string {
	return r.destContentType
}

// TODO create Key() and Hash() funcs that will give a key (finalURL) and hash
// for checking against duplicates

func harvestResource(h *ContentHarvester, origURLtext string) *HarvestedResource {
	result := new(HarvestedResource)
	result.origURLtext = origURLtext

	// Use the standard Go HTTP library method to retrieve the content; the
	// default will automatically follow redirects (e.g. HTTP redirects)
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

	// If we get to here, it means that the URL is valid and the destination is real.
	// But, it doesnt mean there are no other redirects.
	// TODO study other types: http://redirectdetective.com/redirection-types.html
	// TODO add Meta Refresh detection, try it out on this URL:
	//      https://t.co/4dcdNEQYHa, which redirects to http://okt.to/7QDUYM,
	//      which uses meta refresh to redirect but the code in this package doesn't
	//      handle the redirect

	result.destContentType = resp.Header.Get("Content-type")
	result.resolvedURL = resp.Request.URL
	result.finalURL = result.resolvedURL
	ignoreURL, ignoreReason := h.ignoreResource(result.resolvedURL)
	if ignoreURL {
		result.isDestValid = true
		result.isURLIgnored = true
		result.ignoreReason = ignoreReason
		return result
	}

	result.isURLIgnored = false
	result.isDestValid = true
	urlsParamsCleaned, cleanedURL := h.cleanResource(result.resolvedURL)
	if urlsParamsCleaned {
		result.cleanedURL = cleanedURL
		result.finalURL = cleanedURL
		result.isURLCleaned = true
	} else {
		result.isURLCleaned = false
	}

	// TODO once the URL is cleaned, double-check the cleaned URL to see if it's a valid destination; if not, revert to non-cleaned version
	// this could be done recursively here or by the outer function

	return result
}
