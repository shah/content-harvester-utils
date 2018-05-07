package harvester

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// metaRefreshContentRegEx is used to match the 'content' attribute in a tag like this:
//   <meta http-equiv="refresh" content="2;url=https://www.google.com">
var metaRefreshContentRegEx = regexp.MustCompile(`^(\d?)\s?;\s?url=(.*)$`)

// HarvestedResource tracks a single URL that was discovered in content.
// Discovered URLs are validated, follow their redirects, and may have
// query parameters "cleaned" (if instructed).
// TODO need to add heavy automated testing through unit tests
type HarvestedResource struct {
	// TODO consider adding source information (e.g. tweet, e-mail, etc.) and embed style (e.g. text, HTML <a> tag, etc.)
	origURLtext     string
	origResource    *HarvestedResource
	isURLValid      bool
	isDestValid     bool
	httpStatusCode  int
	isURLIgnored    bool
	ignoreReason    string
	isURLCleaned    bool
	destContentType string
	isHTMLRedirect  bool
	htmlRedirectURL string
	htmlParseError  error
	resolvedURL     *url.URL
	cleanedURL      *url.URL
	finalURL        *url.URL
}

// OriginalURLText returns the URL as it was discovered, with no alterations
func (r *HarvestedResource) OriginalURLText() string {
	return r.origURLtext
}

// ReferredByResource returns the original resource that referred this one,
// which is only non-nil when this resource was an HTML (not HTTP) redirect
func (r *HarvestedResource) ReferredByResource() *HarvestedResource {
	return r.origResource
}

// IsValid indicates whether (a) the original URL was parseable and (b) whether
// the destination is valid -- meaning not a 404 or something else
func (r *HarvestedResource) IsValid() (bool, bool) {
	return r.isURLValid, r.isDestValid
}

// IsIgnored indicates whether the URL should be ignored based on harvesting rules.
// Discovered URLs may be ignored for a variety of reasons using a list of Regexps.
func (r *HarvestedResource) IsIgnored() (bool, string) {
	return r.isURLIgnored, r.ignoreReason
}

// IsCleaned indicates whether URL query parameters were removed and the new "cleaned" URL
func (r *HarvestedResource) IsCleaned() (bool, *url.URL) {
	return r.isURLCleaned, r.cleanedURL
}

// GetURLs returns the final (most useful), originally resolved, and "cleaned" URLs
func (r *HarvestedResource) GetURLs() (*url.URL, *url.URL, *url.URL) {
	return r.finalURL, r.resolvedURL, r.cleanedURL
}

// IsHTMLRedirect returns true if redirect was requested through via <meta http-equiv='refresh' content='delay;url='>
// For an explanation, please see http://redirectdetective.com/redirection-types.html
func (r *HarvestedResource) IsHTMLRedirect() (bool, string) {
	return r.isHTMLRedirect, r.htmlRedirectURL
}

// DestinationContentType returns the MIME type of the destination
func (r *HarvestedResource) DestinationContentType() string {
	return r.destContentType
}

// TODO create Key() and Hash() funcs that will give a key (finalURL) and hash
// for checking against duplicates

func cleanResource(url *url.URL, rule CleanDiscoveredResourceRule) (bool, *url.URL) {
	if !rule.CleanDiscoveredResource(url) {
		return false, nil
	}

	// make a copy because we're planning on changing the URL params
	cleanedURL, error := url.Parse(url.String())
	if error != nil {
		return false, nil
	}

	harvestedParams := cleanedURL.Query()
	type ParamMatch struct {
		paramName string
		reason    string
	}
	var cleanedParams []ParamMatch
	for paramName := range harvestedParams {
		remove, reason := rule.RemoveQueryParamFromResource(paramName)
		if remove {
			harvestedParams.Del(paramName)
			cleanedParams = append(cleanedParams, ParamMatch{paramName, reason})
		}
	}

	if len(cleanedParams) > 0 {
		cleanedURL.RawQuery = harvestedParams.Encode()
		return true, cleanedURL
	}
	return false, nil
}

func findMetaRefreshTagInHead(doc *html.Node) *html.Node {
	var metaTag *html.Node
	var inHead bool
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.EqualFold(n.Data, "head") {
			inHead = true
		}
		if inHead && n.Type == html.ElementNode && strings.EqualFold(n.Data, "meta") {
			for _, attr := range n.Attr {
				if strings.EqualFold(attr.Key, "http-equiv") && strings.EqualFold(strings.TrimSpace(attr.Val), "refresh") {
					metaTag = n
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return metaTag
}

// See for explanation: http://redirectdetective.com/redirection-types.html
func getMetaRefresh(resp *http.Response) (bool, string, error) {
	doc, parseError := html.Parse(resp.Body)
	if parseError != nil {
		return false, "", parseError
	}
	defer resp.Body.Close()

	mn := findMetaRefreshTagInHead(doc)
	if mn == nil {
		return false, "", nil
	}

	for _, attr := range mn.Attr {
		if strings.EqualFold(attr.Key, "content") {
			contentValue := strings.TrimSpace(attr.Val)
			parts := metaRefreshContentRegEx.FindStringSubmatch(contentValue)
			if parts != nil && len(parts) == 3 {
				// the first part is the entire match
				// the second and third parts are the delay and URL
				return true, parts[2], nil
			}
		}
	}

	return false, "", nil
}

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

	result.httpStatusCode = resp.StatusCode
	if result.httpStatusCode != 200 {
		result.isDestValid = false
		result.isURLIgnored = true
		result.ignoreReason = fmt.Sprintf("Invalid HTTP Status Code %d", resp.StatusCode)
		return result
	}

	result.destContentType = resp.Header.Get("Content-type")
	result.resolvedURL = resp.Request.URL
	result.finalURL = result.resolvedURL
	ignoreURL, ignoreReason := h.ignoreResourceRule.IgnoreDiscoveredResource(result.resolvedURL)
	if ignoreURL {
		result.isDestValid = true
		result.isURLIgnored = true
		result.ignoreReason = ignoreReason
		return result
	}

	result.isURLIgnored = false
	result.isDestValid = true
	urlsParamsCleaned, cleanedURL := cleanResource(result.resolvedURL, h.cleanResourceRule)
	if urlsParamsCleaned {
		result.cleanedURL = cleanedURL
		result.finalURL = cleanedURL
		result.isURLCleaned = true
	} else {
		result.isURLCleaned = false
	}

	result.isHTMLRedirect, result.htmlRedirectURL, result.htmlParseError = getMetaRefresh(resp)

	// TODO once the URL is cleaned, double-check the cleaned URL to see if it's a valid destination; if not, revert to non-cleaned version
	// this could be done recursively here or by the outer function. This is necessary because "cleaning" a URL and removing params might
	// break it so we need to revert to original.

	return result
}

func harvestResourceFromReferrer(h *ContentHarvester, original *HarvestedResource) *HarvestedResource {
	isHTMLRedirect, htmlRedirectURL := original.IsHTMLRedirect()
	if !isHTMLRedirect {
		return nil
	}

	result := harvestResource(h, htmlRedirectURL)
	result.origResource = original
	return result
}
