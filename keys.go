package harvester

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Machiel/slugify"
	"github.com/julianshen/og"
)

// HarvestedResourceKeys tracks the keys for a single URL that was discovered in content.
// Keys allow the URL to be identified in a database, key value store, etc.
type HarvestedResourceKeys struct {
	hr       *HarvestedResource
	uniqueID string
	pageInfo *og.PageInfo
	piError  error
}

// UniqueID returns the unique identifier based on key searching algorithm
func (keys *HarvestedResourceKeys) UniqueID() string {
	return keys.uniqueID
}

// IsValid returns true if there are no errors
func (keys *HarvestedResourceKeys) IsValid() bool {
	return keys.piError == nil
}

// Slug returns the title of the content
func (keys *HarvestedResourceKeys) Slug() string {
	if keys.piError == nil {
		return slugify.Slugify(keys.pageInfo.Title)
	}
	return "Error getting PageInfo"
}

// KeyExists is a function passed in that checks whether a key already exists
type KeyExists func(key string, random uint32, try int) bool

// GenerateUniqueID generates a unique identifier for this resource
func generateUniqueID(prefix string, existsFn KeyExists) string {
	nconflict := 0
	for i := 0; i < 10000; i++ {
		nextInt, nextSuffix := nextSuffix()
		name := prefix + nextSuffix
		if !existsFn(name, nextInt, i) {
			return name
		}

		if nconflict++; nconflict > 10 {
			randmu.Lock()
			rand = reseed()
			randmu.Unlock()
		}
	}

	// give up after max tries, not much we can do
	_, nextSuffix := nextSuffix()
	return prefix + nextSuffix
}

// CreateKeys returns a new resource keys object
func CreateKeys(hr *HarvestedResource, uniqueIDPrefix string, existsFn KeyExists) *HarvestedResourceKeys {
	result := new(HarvestedResourceKeys)
	result.hr = hr
	result.uniqueID = generateUniqueID(uniqueIDPrefix, existsFn)
	// TODO this does an extra HTTP get, instead we should re-use a downloaded HTML
	result.pageInfo, result.piError = og.GetPageInfoFromUrl(hr.finalURL.String())
	return result
}

// Random number state, approach copied from tempfile.go standard library
var rand uint32
var randmu sync.Mutex

func reseed() uint32 {
	return uint32(time.Now().UnixNano() + int64(os.Getpid()))
}

func nextSuffix() (uint32, string) {
	randmu.Lock()
	r := rand
	if r == 0 {
		r = reseed()
	}
	r = r*1664525 + 1013904223 // constants from Numerical Recipes
	rand = r
	randmu.Unlock()
	next := 1e9 + r%1e9
	return next, strconv.Itoa(int(next))[1:]
}
