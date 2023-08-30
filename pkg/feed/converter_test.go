package feed

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"
	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sampleNitterFeed = gofeed.Feed{
	Title:           "Coldplay / @coldplay",
	Description:     "Twitter feed for: @coldplay. Generated by nitter.moomoo.me",
	Link:            "http://nitter.moomoo.me/coldplay",
	FeedLink:        "https://nitter.moomoo.me/coldplay/rss",
	Links:           []string{"http://nitter.moomoo.me/coldplay"},
	PublishedParsed: &actualTime,
	Language:        "en-us",
	Image: &gofeed.Image{
		URL:   "http://nitter.moomoo.me/pic/pbs.twimg.com%2Fprofile_images%2F1417506973877211138%2FYIm7dOQH_400x400.jpg",
		Title: "Coldplay / @coldplay",
	},
}

var sampleStackerNewsFeed = gofeed.Feed{
	Title:           "Stacker News",
	Description:     "Like Hacker News, but we pay you Bitcoin.",
	Link:            "https://stacker.news",
	FeedLink:        "https://stacker.news/rss",
	Links:           []string{"https://blog.cryptographyengineering.com/2014/11/zero-knowledge-proofs-illustrated-primer.html"},
	PublishedParsed: &actualTime,
	Language:        "en",
}

var sampleNitterFeedRTItem = gofeed.Item{
	Title:           "RT by @coldplay: TOMORROW",
	Description:     "Sample description",
	Content:         "Sample content",
	Link:            "http://nitter.moomoo.me/coldplay/status/1622148481740685312#m",
	UpdatedParsed:   &actualTime,
	PublishedParsed: &actualTime,
	GUID:            "http://nitter.moomoo.me/coldplay/status/1622148481740685312#m",
	DublinCoreExt: &ext.DublinCoreExtension{
		Creator: []string{"@nbcsnl"},
	},
}

var sampleNitterFeedResponseItem = gofeed.Item{
	Title:           "R to @coldplay: Sample",
	Description:     "Sample description",
	Content:         "Sample content",
	Link:            "http://nitter.moomoo.me/elonmusk/status/1621544996167122944#m",
	UpdatedParsed:   &actualTime,
	PublishedParsed: &actualTime,
	GUID:            "http://nitter.moomoo.me/elonmusk/status/1621544996167122944#m",
	DublinCoreExt: &ext.DublinCoreExtension{
		Creator: []string{"@elonmusk"},
	},
}

var sampleDefaultFeedItem = gofeed.Item{
	Title:           "Golang Weekly",
	Description:     "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Phasellus nec condimentum orci. Vestibulum at nunc porta, placerat ex sit amet, consectetur augue. Donec cursus ipsum sed venenatis maximus. Nunc tincidunt dui nec congue lacinia. In mollis magna eu nisi viverra luctus. Ut ultrices eros gravida, lacinia nibh vitae, tristique massa. Sed eu scelerisque erat. Sed eget tortor et turpis feugiat interdum. Nulla sit amet nibh vel massa bibendum congue. Quisque sed tempor velit. Interdum et malesuada fames ac ante ipsum primis in faucibus. Curabitur suscipit mollis fringilla. Integer quis sodales tortor, at hendrerit lacus. Cras posuere maximus nisi. Mauris eget.",
	Content:         "Sample content",
	Link:            "https://golangweekly.com/issues/446",
	UpdatedParsed:   &actualTime,
	PublishedParsed: &actualTime,
	GUID:            "https://golangweekly.com/issues/446",
}

var sampleDefaultFeedItemWithComments = gofeed.Item{
	Title:           "Golang Weekly",
	Description:     "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Phasellus nec condimentum orci. Vestibulum at nunc porta, placerat ex sit amet, consectetur augue. Donec cursus ipsum sed venenatis maximus. Nunc tincidunt dui nec congue lacinia. In mollis magna eu nisi viverra luctus. Ut ultrices eros gravida, lacinia nibh vitae, tristique massa. Sed eu scelerisque erat. Sed eget tortor et turpis feugiat interdum. Nulla sit amet nibh vel massa bibendum congue. Quisque sed tempor velit. Interdum et malesuada fames ac ante ipsum primis in faucibus. Curabitur suscipit mollis fringilla. Integer quis sodales tortor, at hendrerit lacus. Cras posuere maximus nisi. Mauris eget.",
	Content:         "Sample content",
	Link:            "https://golangweekly.com/issues/446",
	UpdatedParsed:   &actualTime,
	PublishedParsed: &actualTime,
	GUID:            "https://golangweekly.com/issues/446",
	Custom: map[string]string{
		"comments": "https://golangweekly.com/issues/446",
	},
}

var sampleDefaultFeedItemExpectedContent = fmt.Sprintf("**%s**\n\n%s", sampleDefaultFeedItem.Title, sampleDefaultFeedItem.Description)
var sampleDefaultFeedItemExpectedContentSubstring = sampleDefaultFeedItemExpectedContent[0:249]

var sampleStackerNewsFeedItem = gofeed.Item{
	Title:           "Zero Knowledge Proofs: An illustrated primer",
	Description:     "<a href=\"https://stacker.news/items/131533\">Comments</a>",
	Content:         "Sample content",
	Link:            "https://blog.cryptographyengineering.com/2014/11/zero-knowledge-proofs-illustrated-primer.html",
	UpdatedParsed:   &actualTime,
	PublishedParsed: &actualTime,
	GUID:            "https://stacker.news/items/131533",
	Custom: map[string]string{
		"comments": "https://stacker.news/items/131533",
	},
}

var sampleDefaultFeed = gofeed.Feed{
	Title:           "Golang Weekly",
	Description:     "A weekly newsletter about the Go programming language",
	Link:            "https://golangweekly.com/rss",
	FeedLink:        "https://golangweekly.com/rss",
	Links:           []string{"https://golangweekly.com/issues/446"},
	PublishedParsed: &actualTime,
	Language:        "en-us",
	Image:           nil,
}

func TestNoteConverter(t *testing.T) {
	testCases := []struct {
		pubKey           string
		item             *gofeed.Item
		feed             *gofeed.Feed
		defaultCreatedAt time.Time
		originalUrl      string
		expectedContent  string
		maxContentLength int
		expectedTags     nostr.Tags
	}{
		{
			pubKey:           samplePubKey,
			item:             &sampleNitterFeedRTItem,
			feed:             &sampleNitterFeed,
			defaultCreatedAt: actualTime,
			originalUrl:      sampleNitterFeed.FeedLink,
			expectedContent:  fmt.Sprintf("**RT %s:**\n\n%s\n\n%s", sampleNitterFeedRTItem.DublinCoreExt.Creator[0], sampleNitterFeedRTItem.Description, strings.ReplaceAll(sampleNitterFeedRTItem.Link, "http://", "https://")),
			maxContentLength: 250,
			expectedTags: nostr.Tags{
				nostr.Tag{"proxy", "https://nitter.moomoo.me/coldplay/rss#http%3A%2F%2Fnitter.moomoo.me%2Fcoldplay%2Fstatus%2F1622148481740685312%23m", "rss"},
			},
		},
		{
			pubKey:           samplePubKey,
			item:             &sampleNitterFeedResponseItem,
			feed:             &sampleNitterFeed,
			defaultCreatedAt: actualTime,
			originalUrl:      sampleNitterFeed.FeedLink,
			expectedContent:  fmt.Sprintf("**Response to %s:**\n\n%s\n\n%s", "@coldplay", sampleNitterFeedResponseItem.Description, strings.ReplaceAll(sampleNitterFeedResponseItem.Link, "http://", "https://")),
			maxContentLength: 250,
			expectedTags: nostr.Tags{
				nostr.Tag{"proxy", "https://nitter.moomoo.me/coldplay/rss#http%3A%2F%2Fnitter.moomoo.me%2Felonmusk%2Fstatus%2F1621544996167122944%23m", "rss"},
			},
		},
		{
			pubKey:           samplePubKey,
			item:             &sampleDefaultFeedItem,
			feed:             &sampleDefaultFeed,
			defaultCreatedAt: actualTime,
			originalUrl:      sampleDefaultFeed.FeedLink,
			expectedContent:  sampleDefaultFeedItemExpectedContentSubstring + "…" + "\n\n" + sampleDefaultFeedItem.Link,
			maxContentLength: 250,
			expectedTags: nostr.Tags{
				nostr.Tag{"proxy", "https://golangweekly.com/rss#https%3A%2F%2Fgolangweekly.com%2Fissues%2F446", "rss"},
			},
		},
		{
			pubKey:           samplePubKey,
			item:             &sampleDefaultFeedItemWithComments,
			feed:             &sampleDefaultFeed,
			defaultCreatedAt: actualTime,
			originalUrl:      sampleDefaultFeed.FeedLink,
			expectedContent:  sampleDefaultFeedItemExpectedContentSubstring + "…\n\nComments: " + sampleDefaultFeedItemWithComments.Custom["comments"] + "\n\n" + sampleDefaultFeedItem.Link,
			maxContentLength: 250,
			expectedTags: nostr.Tags{
				nostr.Tag{"proxy", "https://golangweekly.com/rss#https%3A%2F%2Fgolangweekly.com%2Fissues%2F446", "rss"},
			},
		},
		{
			pubKey:           samplePubKey,
			item:             &sampleDefaultFeedItemWithComments,
			feed:             &sampleDefaultFeed,
			defaultCreatedAt: actualTime,
			originalUrl:      sampleDefaultFeed.FeedLink,
			expectedContent:  sampleDefaultFeedItemExpectedContent + "\n\nComments: " + sampleDefaultFeedItemWithComments.Custom["comments"] + "\n\n" + sampleDefaultFeedItem.Link,
			maxContentLength: 1500,
			expectedTags: nostr.Tags{
				nostr.Tag{"proxy", "https://golangweekly.com/rss#https%3A%2F%2Fgolangweekly.com%2Fissues%2F446", "rss"},
			},
		},
		{
			pubKey:           samplePubKey,
			item:             &sampleStackerNewsFeedItem,
			feed:             &sampleStackerNewsFeed,
			defaultCreatedAt: actualTime,
			originalUrl:      sampleStackerNewsFeed.FeedLink,
			expectedContent:  fmt.Sprintf("**%s**\n\nComments: %s\n\n%s", sampleStackerNewsFeedItem.Title, sampleStackerNewsFeedItem.GUID, sampleStackerNewsFeedItem.Link),
			maxContentLength: 250,
			expectedTags: nostr.Tags{
				nostr.Tag{"proxy", "https://stacker.news/rss#https%3A%2F%2Fstacker.news%2Fitems%2F131533", "rss"},
			},
		},
	}
	for _, tc := range testCases {
		converter, err := NewNoteConverter(tc.maxContentLength)
		require.NoError(t, err)

		event := converter.Convert(tc.pubKey, tc.item, tc.feed, tc.defaultCreatedAt, tc.originalUrl)
		assert.NotEmpty(t, event)
		assert.Equal(t, tc.pubKey, event.PubKey)
		assert.Equal(t, tc.defaultCreatedAt, event.CreatedAt.Time())
		assert.Equal(t, 1, event.Kind)
		assert.Equal(t, tc.expectedContent, event.Content)
		assert.Empty(t, event.Sig)
		assert.Equal(t, tc.expectedTags, event.Tags)
	}
}

var sampleSubstackFeed = gofeed.Feed{
	Title:           "Yaka Stuff",
	Description:     "News, industry perspectives, and updates from the Hard Yaka ecosystem.",
	Link:            "https://hardyaka.substack.com",
	FeedLink:        "https://hardyaka.substack.com",
	Links:           []string{"https://hardyaka.substack.com"},
	PublishedParsed: &actualTime,
	Language:        "en",
	Image: &gofeed.Image{
		URL:   "https://substackcdn.com/image/fetch/w_256,c_limit,f_auto,q_auto:good,fl_progressive:steep/https%3A%2F%2Fhardyaka.substack.com%2Fimg%2Fsubstack.png",
		Title: "Yaka Stuff",
	},
}

var sampleSubstackFeedItem = gofeed.Item{
	Title:           "This is the Universal Ledger",
	Description:     "A core part of the Hard Yaka ecosystem vision",
	Content:         "<div class=\"captioned-image-container\"><figure><a class=\"image-link is-viewable-img image2\" target=\"_blank\" href=\"https://substackcdn.com/image/fetch/f_auto,q_auto:good,fl_progressive:steep/https%3A%2F%2Fsubstack-post-media.s3.amazonaws.com%2Fpublic%2Fimages%2F8366d7e8-05b5-4686-80a8-7648c60d923f_3557x2000.png\"><div class=\"image2-inset\"><picture><source type=\"image/webp\" srcset=\"https://substackcdn.com/image/fetch/w_424,c_limit,f_webp,q_auto:good,fl_progressive:steep/https%3A%2F%2Fsubstack-post-media.s3.amazonaws.com%2Fpublic%2Fimages%2F8366d7e8-05b5-4686-80a8-7648c60d923f_3557x2000.png 424w, https://substackcdn.com/image/fetch/w_848,c_limit,f_webp,q_auto:good,fl_progressive:steep/https%3A%2F%2Fsubstack-post-media.s3.amazonaws.com%2Fpublic%2Fimages%2F8366d7e8-05b5-4686-80a8-7648c60d923f_3557x2000.png 848w, https://substackcdn.com/image/fetch/w_1272,c_limit,f_webp,q_auto:good,fl_progressive:steep/https%3A%2F%2Fsubstack-post-media.s3.amazonaws.com%2Fpublic%2Fimages%2F8366d7e8-05b5-4686-80a8-7648c60d923f_3557x2000.png 1272w, https://substackcdn.com/image/fetch/w_1456,c_limit,f_webp,q_auto:good,fl_progressive:steep/https%3A%2F%2Fsubstack-post-media.s3.amazonaws.com%2Fpublic%2Fimages%2F8366d7e8-05b5-4686-80a8-7648c60d923f_3557x2000.png 1456w\" sizes=\"100vw\"><img src=\"https://substackcdn.com/image/fetch/w_1456,c_limit,f_auto,q_auto:good,fl_progressive:steep/https%3A%2F%2Fsubstack-post-media.s3.amazonaws.com%2Fpublic%2Fimages%2F8366d7e8-05b5-4686-80a8-7648c60d923f_3557x2000.png\" width=\"1456\" height=\"819\" data-attrs=\"{&quot;src&quot;:&quot;https://substack-post-media.s3.amazonaws.com/public/images/8366d7e8-05b5-4686-80a8-7648c60d923f_3557x2000.png&quot;,&quot;fullscreen&quot;:null,&quot;imageSize&quot;:null,&quot;height&quot;:819,&quot;width&quot;:1456,&quot;resizeWidth&quot;:null,&quot;bytes&quot;:5096872,&quot;alt&quot;:null,&quot;title&quot;:null,&quot;type&quot;:&quot;image/png&quot;,&quot;href&quot;:null,&quot;belowTheFold&quot;:false,&quot;internalRedirect&quot;:null}\" class=\"sizing-normal\" alt=\"\" srcset=\"https://substackcdn.com/image/fetch/w_424,c_limit,f_auto,q_auto:good,fl_progressive:steep/https%3A%2F%2Fsubstack-post-media.s3.amazonaws.com%2Fpublic%2Fimages%2F8366d7e8-05b5-4686-80a8-7648c60d923f_3557x2000.png 424w, https://substackcdn.com/image/fetch/w_848,c_limit,f_auto,q_auto:good,fl_progressive:steep/https%3A%2F%2Fsubstack-post-media.s3.amazonaws.com%2Fpublic%2Fimages%2F8366d7e8-05b5-4686-80a8-7648c60d923f_3557x2000.png 848w, https://substackcdn.com/image/fetch/w_1272,c_limit,f_auto,q_auto:good,fl_progressive:steep/https%3A%2F%2Fsubstack-post-media.s3.amazonaws.com%2Fpublic%2Fimages%2F8366d7e8-05b5-4686-80a8-7648c60d923f_3557x2000.png 1272w, https://substackcdn.com/image/fetch/w_1456,c_limit,f_auto,q_auto:good,fl_progressive:steep/https%3A%2F%2Fsubstack-post-media.s3.amazonaws.com%2Fpublic%2Fimages%2F8366d7e8-05b5-4686-80a8-7648c60d923f_3557x2000.png 1456w\" sizes=\"100vw\"></picture><div class=\"image-link-expand\"><svg xmlns=\"http://www.w3.org/2000/svg\" width=\"16\" height=\"16\" viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"#FFFFFF\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\" class=\"lucide lucide-maximize2\"><polyline points=\"15 3 21 3 21 9\"></polyline><polyline points=\"9 21 3 21 3 15\"></polyline><line x1=\"21\" y1=\"3\" x2=\"14\" y2=\"10\"></line><line x1=\"3\" y1=\"21\" x2=\"10\" y2=\"14\"></line></svg></div></div></a></figure></div><p>This morning, <a href=\"https://uled.io/\">the Universal Ledger</a> came out of stealth mode, having raised $10 million in a seed round from Hard Yaka.</p><p>Universal Ledger is a core part of the Hard Yaka ecosystem vision, bringing together tokenized U.S. dollar assets, progressive digital identity, and compliance baked directly into a permissioned blockchain.</p><p>The end result is a digital wallet platform that allows businesses and governments to develop wallet experiences with global U.S. dollar access, real-time payments, and 24/7 availability at a fraction of the cost.</p><ul><li><p><strong><a href=\"https://medium.com/universal-ledger/introducing-the-universal-ledger-5f87aa07fb74\">Introducing the Universal Ledger</a></strong></p></li></ul><p>From the press release:</p><blockquote><p><em>Universal Ledger provides digital asset and payment infrastructure on an identity and compliance-first blockchain. It is the first blockchain platform to apply digital identity to a fully compliant infrastructure, with the ability to control digital assets.</em></p><p><em>Universal Ledger is a true wallet-as-a-service platform, providing developers and engineers with an API and event-driven architecture on which they can build digital wallet experiences that allow assets to be moved globally, in real time, while maintaining global compliance standards and local jurisdictional identity requirements.</em></p></blockquote><p>Universal Ledger is helmed by Kirk Chapman, who has 20 years of experience in the payments, banking, and fintech industries, having served as Head of Strategy at Galileo and was advisor to the CEO at SoFi.</p><ul><li><p><strong><a href=\"https://medium.com/universal-ledger/a-conversation-with-universal-ledger-co-founder-and-ceo-kirk-chapman-7226df068072\">A conversation with Universal Ledger co-founder and CEO Kirk Chapman</a></strong></p></li></ul><p>Here&#8217;s Kirk:</p><blockquote><p><em>We&#8217;re providing a way for companies to build digital wallet experiences such that they can have peace of mind when it comes to compliance and custody. Universal Ledger fundamentally changes the payments model by leveraging blockchain technology and digital identity, eliminating substantial abuse and fraud. This allows businesses and governments to not only explore new use cases and more efficient disbursement models but also bring previously unreachable people into the contemporary global financial system.</em></p></blockquote><p>And here&#8217;s Hard Yaka partner, Greg Kidd:</p><blockquote><p><em>The Universal Ledger heralds the day when anyone with a baseline identity can hold, send, and receive dollars safely and compliantly. Site and app developers can utilize Universal Ledger&#8217;s APIs to include balance payment functions in their offerings while minimizing their regulatory burden because risk and compliance controls are built directly into the ledger itself rather than maintained at the client or wallet level. Balances are global, interoperable, and operate 24/7&#8212;powered by blockchain technology, but without the need for traditional crypto fees or tokens.</em></p></blockquote><p>The Universal Ledger launch was <a href=\"https://www.axios.com/pro/fintech-deals/2023/04/19/universal-ledger-10m-wallet-as-a-service\">covered by Axios</a>.</p><p>For media inquiries, please reach out to <a href=\"mailto:roger@methodcommunications.com\">Roger Johnson</a>.</p><p><em><strong>Learn more about <a href=\"https://uled.io/\">Universal Ledger</a></strong></em></p>",
	Link:            "https://hardyaka.substack.com/p/this-is-the-universal-ledger",
	UpdatedParsed:   &actualTime,
	PublishedParsed: &actualTime,
	GUID:            "https://hardyaka.substack.com/p/this-is-the-universal-ledger",
}

var expectedSampleSubstackFeedItemEventContent = `**This is the Universal Ledger**

![](https://substackcdn.com/image/fetch/w_1456,c_limit,f_auto,q_auto:good,fl_progressive:steep/https%3A%2F%2Fsubstack-post-media.s3.amazonaws.com%2Fpublic%2Fimages%2F8366d7e8-05b5-4686-80a8-7648c60d923f_3557x2000.png) (https://substackcdn.com/image/fetch/f_auto,q_auto:good,fl_progressive:steep/https%3A%2F%2Fsubstack-post-media.s3.amazonaws.com%2Fpublic%2Fimages%2F8366d7e8-05b5-4686-80a8-7648c60d923f_3557x2000.png)

This morning, the Universal Ledger (https://uled.io/) came out of stealth mode, having raised $10 million in a seed round from Hard Yaka.

Universal Ledger is a core part of the Hard Yaka ecosystem vision, bringing together tokenized U.S. dollar assets, progressive digital identity, and compliance baked directly into a permissioned blockchain.

The end result is a digital wallet platform that allows businesses and governments to develop wallet experiences with global U.S. dollar access, real-time payments, and 24/7 availability at a fraction of the cost.

- **Introducing the Universal Ledger (https://medium.com/universal-ledger/introducing-the-universal-ledger-5f87aa07fb74)**


From the press release:

> _Universal Ledger provides digital asset and payment infrastructure on an identity and compliance-first blockchain. It is the first blockchain platform to apply digital identity to a fully compliant infrastructure, with the ability to control digital assets._
>
> _Universal Ledger is a true wallet-as-a-service platform, providing developers and engineers with an API and event-driven architecture on which they can build digital wallet experiences that allow assets to be moved globally, in real time, while maintaining global compliance standards and local jurisdictional identity requirements._

Universal Ledger is helmed by Kirk Chapman, who has 20 years of experience in the payments, banking, and fintech industries, having served as Head of Strategy at Galileo and was advisor to the CEO at SoFi.

- **A conversation with Universal Ledger co-founder and CEO Kirk Chapman (https://medium.com/universal-ledger/a-conversation-with-universal-ledger-co-founder-and-ceo-kirk-chapman-7226df068072)**


Here’s Kirk:

> _We’re providing a way for companies to build digital wallet experiences such that they can have peace of mind when it comes to compliance and custody. Universal Ledger fundamentally changes the payments model by leveraging blockchain technology and digital identity, eliminating substantial abuse and fraud. This allows businesses and governments to not only explore new use cases and more efficient disbursement models but also bring previously unreachable people into the contemporary global financial system._

And here’s Hard Yaka partner, Greg Kidd:

> _The Universal Ledger heralds the day when anyone with a baseline identity can hold, send, and receive dollars safely and compliantly. Site and app developers can utilize Universal Ledger’s APIs to include balance payment functions in their offerings while minimizing their regulatory burden because risk and compliance controls are built directly into the ledger itself rather than maintained at the client or wallet level. Balances are global, interoperable, and operate 24/7—powered by blockchain technology, but without the need for traditional crypto fees or tokens._

The Universal Ledger launch was covered by Axios (https://www.axios.com/pro/fintech-deals/2023/04/19/universal-ledger-10m-wallet-as-a-service).

For media inquiries, please reach out to Roger Johnson (mailto:roger@methodcommunications.com).

_**Learn more about Universal Ledger (https://uled.io/)**_

https://hardyaka.substack.com/p/this-is-the-universal-ledger`

func TestLongFormConverter(t *testing.T) {
	testCases := []struct {
		pubKey           string
		item             *gofeed.Item
		feed             *gofeed.Feed
		defaultCreatedAt time.Time
		originalUrl      string
		expectedContent  string
		expectedTags     nostr.Tags
	}{
		{
			pubKey:           samplePubKey,
			item:             &sampleSubstackFeedItem,
			feed:             &sampleSubstackFeed,
			defaultCreatedAt: actualTime,
			originalUrl:      sampleSubstackFeed.FeedLink,
			expectedContent:  expectedSampleSubstackFeedItemEventContent,
			expectedTags: nostr.Tags{
				nostr.Tag{"published_at", strconv.FormatInt(actualTime.Unix(), 10)},
				nostr.Tag{"d", "https://hardyaka.substack.com/p/this-is-the-universal-ledger"},
				nostr.Tag{"title", "This is the Universal Ledger"},
				nostr.Tag{"proxy", "https://hardyaka.substack.com#https%3A%2F%2Fhardyaka.substack.com%2Fp%2Fthis-is-the-universal-ledger", "rss"},
			},
		},
	}
	for _, tc := range testCases {
		converter := NewLongFormConverter()

		event := converter.Convert(tc.pubKey, tc.item, tc.feed, tc.defaultCreatedAt, tc.originalUrl)
		assert.NotEmpty(t, event)
		assert.Equal(t, tc.pubKey, event.PubKey)
		assert.Equal(t, tc.defaultCreatedAt, event.CreatedAt.Time())
		assert.Equal(t, 30023, event.Kind)
		assert.Equal(t, tc.expectedContent, event.Content)
		assert.Empty(t, event.Sig)
		assert.Equal(t, tc.expectedTags, event.Tags)
	}
}
