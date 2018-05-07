package harvester

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ResourceSuite struct {
	suite.Suite
	ch *ContentHarvester
}

func (suite *ResourceSuite) SetupSuite() {
	suite.ch = MakeContentHarvester(defaultIgnoreURLsRegExList, defaultCleanURLsRegExList, false)
}

func (suite *ResourceSuite) harvestSingleURLFromMockTweet(text string, msgAndArgs ...interface{}) *HarvestedResource {
	harvested := suite.ch.HarvestResources(fmt.Sprintf(text, msgAndArgs))
	suite.Equal(len(harvested.Resources), 1)
	return harvested.Resources[0]
}

// TODO remove the 'x' to test this function, it takes a long time to run (need to debug why)
func (suite *ResourceSuite) TestInvalidlyFormattedURLs() {
	hr := suite.harvestSingleURLFromMockTweet("Test an invalidly formatted URL %s in a mock tweet", "https://t")
	isURLValid, isDestValid := hr.IsValid()
	suite.False(isURLValid, "URL should have invalid format")
	suite.False(isDestValid, "URL should have invalid destination")
}

func (suite *ResourceSuite) TestInvalidDestinationURLs() {
	hr := suite.harvestSingleURLFromMockTweet("Test a validly formatted URL %s but with invalid destination in a mock tweet", "https://t.co/fDxPF")
	isURLValid, isDestValid := hr.IsValid()
	suite.True(isURLValid, "URL should be formatted validly")
	suite.False(isDestValid, "URL should have invalid destination")
	suite.Equal(hr.httpStatusCode, 404)
}

func (suite *ResourceSuite) TestIgnoreRules() {
	hr := suite.harvestSingleURLFromMockTweet("Test a good URL %s which will redirect to a URL we want to ignore", "https://t.co/xNzrxkHE1u")
	isURLValid, isDestValid := hr.IsValid()
	suite.True(isURLValid, "URL should be formatted validly")
	suite.True(isDestValid, "URL should have valid destination")
	isIgnored, ignoreReason := hr.IsIgnored()
	suite.True(isIgnored, "URL should be ignored (skipped)")
	suite.Equal(ignoreReason, "Matched Ignore Rule `^https://twitter.com/(.*?)/status/(.*)$`")
}

func (suite *ResourceSuite) TestResolvedURLRedirectedThroughHTMLProperly() {
	hr := suite.harvestSingleURLFromMockTweet("Test a good URL %s which will redirect to a URL we want to resolve via <meta http-equiv='refresh' content='delay;url='>, with utm_* params", "https://t.co/4dcdNEQYHa")
	isURLValid, isDestValid := hr.IsValid()
	suite.True(isURLValid, "URL should be formatted validly")
	suite.True(isDestValid, "URL should have valid destination")
	isIgnored, _ := hr.IsIgnored()
	suite.False(isIgnored, "URL should not be ignored")
	isHTMLRedirect, htmlRedirectURLText := hr.IsHTMLRedirect()
	suite.True(isHTMLRedirect, "There should have been an HTML redirect requested through <meta http-equiv='refresh' content='delay;url='>")
	suite.Equal(htmlRedirectURLText, "https://www.sopranodesign.com/secure-healthcare-messaging/?utm_source=twitter&utm_medium=socialmedia&utm_campaign=soprano")

	// at this point we want to get the "new" (redirected) and test it
	redirectedHR := harvestResourceFromReferrer(suite.ch, hr)
	suite.Equal(redirectedHR.ReferredByResource(), hr, "The referral resource should be the same as the original")
	isURLValid, isDestValid = redirectedHR.IsValid()
	suite.True(isURLValid, "Redirected URL should be formatted validly")
	suite.True(isDestValid, "Redirected URL should have valid destination")
	isIgnored, _ = redirectedHR.IsIgnored()
	suite.False(isIgnored, "Redirected URL should not be ignored")
	isCleaned, _ := redirectedHR.IsCleaned()
	suite.True(isCleaned, "Redirected URL should be 'cleaned'")
	finalURL, resolvedURL, cleanedURL := redirectedHR.GetURLs()
	suite.Equal(resolvedURL.String(), "https://www.sopranodesign.com/secure-healthcare-messaging/?utm_source=twitter&utm_medium=socialmedia&utm_campaign=soprano")
	suite.Equal(cleanedURL.String(), "https://www.sopranodesign.com/secure-healthcare-messaging/")
	suite.Equal(finalURL.String(), cleanedURL.String(), "finalURL should be same as cleanedURL")
}

func (suite *ResourceSuite) TestResolvedURLCleaned() {
	hr := suite.harvestSingleURLFromMockTweet("Test a good URL %s which will redirect to a URL we want to ignore, with utm_* params", "https://t.co/csWpQq5mbn")
	isURLValid, isDestValid := hr.IsValid()
	suite.True(isURLValid, "URL should be formatted validly")
	suite.True(isDestValid, "URL should have valid destination")
	isIgnored, _ := hr.IsIgnored()
	suite.False(isIgnored, "URL should not be ignored")
	isCleaned, _ := hr.IsCleaned()
	suite.True(isCleaned, "URL should be 'cleaned'")
	finalURL, resolvedURL, cleanedURL := hr.GetURLs()
	suite.Equal(resolvedURL.String(), "https://www.washingtonexaminer.com/chris-matthews-trump-russia-collusion-theory-came-apart-with-comey-testimony/article/2625372?utm_campaign=crowdfire&utm_content=crowdfire&utm_medium=social&utm_source=twitter")
	suite.Equal(cleanedURL.String(), "https://www.washingtonexaminer.com/chris-matthews-trump-russia-collusion-theory-came-apart-with-comey-testimony/article/2625372")
	suite.Equal(finalURL.String(), cleanedURL.String(), "finalURL should be same as cleanedURL")
}

func (suite *ResourceSuite) TestResolvedURLNotCleaned() {
	hr := suite.harvestSingleURLFromMockTweet("Test a good URL %s which will redirect to a URL we want to ignore", "https://t.co/ELrZmo81wI")
	isURLValid, isDestValid := hr.IsValid()
	suite.True(isURLValid, "URL should be formatted validly")
	suite.True(isDestValid, "URL should have valid destination")
	isIgnored, _ := hr.IsIgnored()
	suite.False(isIgnored, "URL should not be ignored")
	isCleaned, _ := hr.IsCleaned()
	suite.False(isCleaned, "URL should not have been 'cleaned'")
	finalURL, resolvedURL, cleanedURL := hr.GetURLs()
	suite.Equal(resolvedURL.String(), "http://www.foxnews.com/lifestyle/2018/04/25/photo-donald-trump-look-alike-in-spain-goes-viral.html")
	suite.Equal(finalURL.String(), resolvedURL.String(), "finalURL should be same as resolvedURL")
	suite.Nil(cleanedURL, "cleanedURL should be empty")
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(ResourceSuite))
}
