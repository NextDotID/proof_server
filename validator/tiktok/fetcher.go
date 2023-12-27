package tiktok

import "regexp"

const (
	FINAL_URL_TEMPLATE = "^https://www\\.tiktok\\.com/@(.+?)/video/(\\d+)"
)
var (
	finalUrlRegexp = regexp.MustCompile(FINAL_URL_TEMPLATE)
)

type OEmbedInfo struct {
	// "version": "1.0",
	Version string `json:"version"`
	// "type": "video",
	Type string `json:"type"`
	// "title": "Scramble up ur name & I’ll try to guess it😍❤️ #foryoupage #petsoftiktok #aesthetic",
	Title string `json:"title"`
	// "author_url": "https://www.tiktok.com/@scout2015",
	AuthorURL string `json:"author_url"`
	// "author_name": "Scout & Suki",
	AuthorName string `json:"author_name"`
	// "width": "100%",
	Width string `json:"width"`
	// "height": "100%",
	Height string `json:"height"`
	// "html": "<blockquote
	Html string `json:"html"`
	// "thumbnail_width": 720,
	ThumbnailWidth int `json:"thumbnail_width"`
	// "thumbnail_height": 1280,
	ThumbnailHeight int `json:"thumbnail_height"`
	// "thumbnail_url": "https://p16.muscdn.com/obj/tos-maliva-p-0068/06kv6rfcesljdjr45ukb0000d844090v0200010605",
	ThumbnailUrl string `json:"thumbnail_url"`
	// "provider_url": "https://www.tiktok.com",
	ProviderUrl string `json:"provider_url"`
	// "provider_name": "TikTok"
	ProviderName string `json:"provider_name"`
}

/// fetchOembedInfo fetches OEmbed card info from TikTok.
func fetchOembedInfo(url string) (OEmbedInfo, error) {
	return OEmbedInfo{}, nil // TODO
}
