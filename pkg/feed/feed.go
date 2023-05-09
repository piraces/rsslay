package feed

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"
	"github.com/mmcdole/gofeed"
	"github.com/nbd-wtf/go-nostr"
	"github.com/piraces/rsslay/pkg/converter"
	"github.com/piraces/rsslay/pkg/helpers"
	"github.com/rif/cache2go"
	"html"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	fp        = gofeed.NewParser()
	feedCache = cache2go.New(512, time.Minute*19)
	client    = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 2 {
				return errors.New("stopped after 2 redirects")
			}
			return nil
		},
		Timeout: 5 * time.Second,
	}
)

type Entity struct {
	PublicKey  string
	PrivateKey string
	URL        string
	Nitter     bool
}

var types = []string{
	"rss+xml",
	"atom+xml",
	"feed+json",
	"text/xml",
	"application/xml",
}

func GetFeedURL(url string) string {
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode >= 300 {
		return ""
	}

	ct := resp.Header.Get("Content-Type")
	for _, typ := range types {
		if strings.Contains(ct, typ) {
			return url
		}
	}

	if strings.Contains(ct, "text/html") {
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return ""
		}

		for _, typ := range types {
			href, _ := doc.Find(fmt.Sprintf("link[type*='%s']", typ)).Attr("href")
			if href == "" {
				continue
			}
			if !strings.HasPrefix(href, "http") && !strings.HasPrefix(href, "https") {
				href, _ = helpers.UrlJoin(url, href)
			}
			return href
		}
	}

	return ""
}

func ParseFeed(url string) (*gofeed.Feed, error) {
	if feed, ok := feedCache.Get(url); ok {
		return feed.(*gofeed.Feed), nil
	}
	fp.RSSTranslator = NewCustomTranslator()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return nil, err
	}

	// cleanup a little so we don't store too much junk
	for i := range feed.Items {
		feed.Items[i].Content = ""
	}
	feedCache.Set(url, feed)

	return feed, nil
}

func EntryFeedToSetMetadata(pubkey string, feed *gofeed.Feed, originalUrl string, enableAutoRegistration bool, defaultProfilePictureUrl string) nostr.Event {
	// Handle Nitter special cases (http schema)
	if strings.Contains(feed.Description, "Twitter feed") {
		if strings.HasPrefix(originalUrl, "https://") {
			feed.Description = strings.ReplaceAll(feed.Description, "http://", "https://")
			feed.Title = strings.ReplaceAll(feed.Title, "http://", "https://")
			feed.Image.URL = strings.ReplaceAll(feed.Image.URL, "http://", "https://")
			feed.Link = strings.ReplaceAll(feed.Link, "http://", "https://")
		}
	}

	metadata := map[string]string{
		"name":  feed.Title + " (RSS Feed)",
		"about": feed.Description + "\n\n" + feed.Link,
	}

	if enableAutoRegistration {
		metadata["nip05"] = fmt.Sprintf("%s@%s", originalUrl, "rsslay.nostr.moe")
	}

	if feed.Image != nil {
		metadata["picture"] = feed.Image.URL
	} else if defaultProfilePictureUrl != "" {
		metadata["picture"] = defaultProfilePictureUrl
	}

	content, _ := json.Marshal(metadata)

	createdAt := time.Unix(time.Now().Unix(), 0)
	if feed.PublishedParsed != nil {
		createdAt = *feed.PublishedParsed
	}

	evt := nostr.Event{
		PubKey:    pubkey,
		CreatedAt: nostr.Timestamp(createdAt.Unix()),
		Kind:      nostr.KindSetMetadata,
		Tags:      nostr.Tags{},
		Content:   string(content),
	}
	evt.ID = string(evt.Serialize())

	return evt
}

func ItemToTextNote(pubkey string, item *gofeed.Item, feed *gofeed.Feed, defaultCreatedAt time.Time, originalUrl string, maxContentLength int) nostr.Event {
	content := ""
	if item.Title != "" {
		content = "**" + item.Title + "**"
	}

	mdConverter := md.NewConverter("", true, nil)
	mdConverter.AddRules(converter.GetConverterRules()...)

	description, err := mdConverter.ConvertString(item.Description)
	if err != nil {
		log.Printf("[WARN] failure to convert description to markdown (defaulting to plain text): %v", err)
		p := bluemonday.StripTagsPolicy()
		description = p.Sanitize(item.Description)
	}

	if !strings.EqualFold(item.Title, description) && !strings.Contains(feed.Link, "stacker.news") {
		content += "\n\n" + description
	}

	shouldUpgradeLinkSchema := false

	// Handle Nitter special cases (duplicates and http schema)
	if strings.Contains(feed.Description, "Twitter feed") {
		content = ""
		shouldUpgradeLinkSchema = true

		if strings.HasPrefix(originalUrl, "https://") {
			description = strings.ReplaceAll(description, "http://", "https://")
		}

		if strings.Contains(item.Title, "RT by @") {
			if len(item.DublinCoreExt.Creator) > 0 {
				content = "**" + "RT " + item.DublinCoreExt.Creator[0] + ":**\n\n"
			}
		} else if strings.Contains(item.Title, "R to @") {
			fields := strings.Fields(item.Title)
			if len(fields) >= 2 {
				replyingToHandle := fields[2]
				content = "**" + "Response to " + replyingToHandle + "**\n\n"
			}
		}
		content += description
	}

	if strings.Contains(feed.Link, "reddit.com") {
		var subreddit = strings.Split(feed.Link, "/r/")
		description += fmt.Sprintf("the hashtag is: #%s", subreddit)
	}

	content = html.UnescapeString(content)
	if len(content) > maxContentLength {
		content = content[0:(maxContentLength-1)] + "…"
	}

	if shouldUpgradeLinkSchema {
		item.Link = strings.ReplaceAll(item.Link, "http://", "https://")
	}

	// Handle comments
	if item.Custom != nil {
		if comments, ok := item.Custom["comments"]; ok {
			content += fmt.Sprintf("\n\nComments: %s", comments)
		}
	}

	content += "\n\n" + item.Link

	createdAt := defaultCreatedAt
	if item.UpdatedParsed != nil {
		createdAt = *item.UpdatedParsed
	}
	if item.PublishedParsed != nil {
		createdAt = *item.PublishedParsed
	}

	evt := nostr.Event{
		PubKey:    pubkey,
		CreatedAt: nostr.Timestamp(createdAt.Unix()),
		Kind:      nostr.KindTextNote,
		Tags:      nostr.Tags{},
		Content:   strings.ToValidUTF8(content, ""),
	}
	evt.ID = string(evt.Serialize())

	return evt
}

func PrivateKeyFromFeed(url string, secret string) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(url))
	r := m.Sum(nil)
	return hex.EncodeToString(r)
}

func DeleteInvalidFeed(url string, db *sql.DB) {
	if _, err := db.Exec(`DELETE FROM feeds WHERE url=?`, url); err != nil {
		log.Printf("[ERROR] failure to delete invalid feed: %v", err)
	} else {
		log.Printf("[DEBUG] deleted invalid feed with url %q", url)
	}
}
