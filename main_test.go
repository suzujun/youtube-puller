package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegexChannelTitle(t *testing.T) {
	t.Parallel()
	bs := []byte(`<link itemprop="url" href="http://www.youtube.com/@kikuchan813"><link itemprop="name" content="アイドル鳥越"></span><script type="application/ld+json" nonce="OomvDtAGKIp-PW5XIn21rA">`)
	res := regexChannelTitle.FindSubmatch(bs)
	assert.Len(t, res, 2)
	assert.Equal(t, "アイドル鳥越", string(res[1]))
}

func TestRegexChannelURL(t *testing.T) {
	t.Parallel()
	bs := []byte(`,"browseEndpoint":{"browseId":"UCp0iCvHGMwyfPHpYq7n2sPw","canonicalBaseUrl":"/channel/UCp0iCvHGMwyfPHpYq7n2sPw"}}}]},"lengthText":{"accessibility":{"accessibilityData"`)
	res := regexChannelURL.FindSubmatch(bs)
	assert.Len(t, res, 3)
	assert.Equal(t, "channel/", string(res[1]))
	assert.Equal(t, "UCp0iCvHGMwyfPHpYq7n2sPw", string(res[2]))
	//
	bs = []byte(`Endpoint":{"browseId":"UCPJCP_fon2mOMfibbBPUOYw","canonicalBaseUrl":"/@SAWAYANGAMES"}}}]},"publishedTimeText":{"simpleText":"1 年前"},"viewCountText":{"simple`)
	res = regexChannelURL.FindSubmatch(bs)
	assert.Len(t, res, 3)
	assert.Equal(t, "@", string(res[1]))
	assert.Equal(t, "SAWAYANGAMES", string(res[2]))
}
