package main

type ctxKey int

// Netlify Library client related
const (
	// NetlifyAPIPorcelainLibraryClientKey uniquely identifies context of netlify porcelain library client
	NetlifyAPIPorcelainLibraryClientKey ctxKey = 1 + iota

	// NetlifyAPIPlumbingLibraryClientKey uniquely identifies context of netlify plumbing library client
	NetlifyAPIPlumbingLibraryClientKey ctxKey = 2 + iota
)

// NetlifyAuthTokenKVIdentifier is used to in suffix with userID to identify key in KV store
const NetlifyAuthTokenKVIdentifier string = "_netlifyToken"

// Netlify specific constants
const (
	// NetlifyAuthURL is auth url to make authentication to netlify
	NetlifyAuthURL = "https://app.netlify.com/authorize"

	// NetlifyTokenURL is token url to get token from netlify
	NetlifyTokenURL = "https://api.netlify.com/oauth/token"

	// NetlifyAPIHost is the base URL for making Netlify API request
	NetlifyAPIHost = "api.netlify.com"

	// NetlifyAPIPath is path attached to baseURL for making Netlify API request
	NetlifyAPIPath = "/api/v1"

	// NetlifyDateLayout is the date format returned by Netlify api for dates
	NetlifyDateLayout = "2006-01-02T15:04:05.000Z"
)

// Netlify Build hook related
const (
	// MattermostNetlifyBuildHookTitle is the title of build hooks created by mattermost
	MattermostNetlifyBuildHookTitle = "Mattermost-Netlify-Build-Hook"

	// MattermostNetlifyBuildHookMessage will be message of all build hook deploys from mattermost
	MattermostNetlifyBuildHookMessage = "triggered by Netlify Bot from Mattermost"
)

// SuccessfullyNetlifyConnectedMessage is posted when /connect command is executed and completed
const SuccessfullyNetlifyConnectedMessage string = "#### Welcome to the Mattermost Netlify Plugin!\n" +
	"You've successfully connected your Mattermost account on Netlify.\n\n" +
	"##### Notifications\n" +
	"Write about how to enable notifications.\n" +
	"##### Slash Commands\n"

const (
	// MarkdownSiteListTableHeader is a table rendered in markdown
	MarkdownSiteListTableHeader string = `
| Name   | URL           | Custom domain | Repository | Branch | Managed by | Last updated |
|:-------|:-------------:|:-------------:|------------|--------|------------|-------------:|`

	// MarkdownDeployListTableHeader is table rendered in markdown to show info regarding site build
	MarkdownDeployListTableHeader string = `
| Sequence |   SHA  | Deployed at |  Deploy ID  |
|:--------:|:-------|------------:|-------------|`
)

// MarkdownSiteListDetailTableHeader is a table rendered in markdown for list detail command
const MarkdownSiteListDetailTableHeader string = `
| Name | ID |
|------|:--:|`

const (
	// ActionDisconnectPlugin is used in Post action to identify disconnect button action
	ActionDisconnectPlugin = "ActionDisconnectPlugin"
	// ActionCancel can be used in any Post action to identify cancel action
	ActionCancel = "ActionCancel"
)
