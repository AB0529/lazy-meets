package main

import (
	"net/url"
	"time"
)

// Config the repersentation of a config file
type Config struct {
	// Leave the number of people left which determins when to leave
	Leave int
	// Skip the window of time between the start of the class, and current time to join
	Skip int
	// Email Gmail login Email
	Email string
	// Password Gmail login password
	Password string
}

// Class the repersentation of a class
type Class struct {
	// Name a friendly name for the class
	Name string
	// Weekdays the weekdays the class takes place
	Weekdays []*time.Weekday
	// MeetURL the URL to join
	MeetURL *url.URL
	// JoinTime the time in which  to join the class
	JoinTime *time.Time
}

// Schedule the repersentation of a created schedule
type Schedule []*Class

type GHAuthor struct {
	Login             string `json:"login"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}

type GHUploader struct {
	Login             string `json:"login"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}

type GHAssets struct {
	URL      string      `json:"url"`
	ID       int         `json:"id"`
	NodeID   string      `json:"node_id"`
	Name     string      `json:"name"`
	Label    interface{} `json:"label"`
	Uploader GHUploader `json:"uploader"`
	ContentType        string    `json:"content_type"`
	State              string    `json:"state"`
	Size               int       `json:"size"`
	DownloadCount      int       `json:"download_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	BrowserDownloadURL string    `json:"browser_download_url"`
}

type GHResp struct {
	URL       string `json:"url"`
	AssetsURL string `json:"assets_url"`
	UploadURL string `json:"upload_url"`
	HTMLURL   string `json:"html_url"`
	ID        int    `json:"id"`
	Author    GHAuthor `json:"author"`
	NodeID          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
	Assets          []GHAssets `json:"assets"`
	TarballURL string `json:"tarball_url"`
	ZipballURL string `json:"zipball_url"`
	Body       string `json:"body"`
}
