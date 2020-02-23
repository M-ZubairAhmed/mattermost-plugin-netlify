package main

type ctxKey int

// NetlifyAPILibraryClientKey uniquely identifies context of netlify library client
const NetlifyAPILibraryClientKey ctxKey = 1 + iota

// NetlifyAuthTokenKVIdentifier is used to in suffix with userID to identify key in KV store
const NetlifyAuthTokenKVIdentifier string = "_netlifyToken"

// NetlifyAuthURL is auth url to make authentication to netlify
const NetlifyAuthURL = "https://app.netlify.com/authorize"

// NetlifyTokenURL is token url to get token from netlify
const NetlifyTokenURL = "https://api.netlify.com/oauth/token"

// NetlifyAPIHost is the base URL for making Netlify API request
const NetlifyAPIHost = "api.netlify.com"

// NetlifyAPIPath is path attached to baseURL for making Netlify API request
const NetlifyAPIPath = "/api/v1"

// NetlifyDateLayout is the date format returned by Netlify api for dates
const NetlifyDateLayout = "2006-01-02T15:04:05.000Z"

// SuccessfullyNetlifyConnectedMessage is posted when /connect command is executed and completed
const SuccessfullyNetlifyConnectedMessage string = "#### Welcome to the Mattermost Netlify Plugin!\n" +
	"You've successfully connected your Mattermost account on Netlify.\n\n" +
	"##### Notifications\n" +
	"Write about how to enable notifications.\n" +
	"##### Slash Commands\n"

// MarkdownSiteListTableHeader is a table rendered in markdown
const MarkdownSiteListTableHeader string = `
| Name   | URL           | Custom domain | Repository | Branch | Managed by | Last updated |
|--------|:-------------:|:-------------:|------------|--------|------------|-------------:|`
