package main

type ctxKey int

// Netlify Library client related
const (
	// NetlifyAPIPorcelainLibraryClientKey uniquely identifies context of netlify porcelain library client
	NetlifyAPIPorcelainLibraryClientKey ctxKey = 1 + iota

	// NetlifyAPIPlumbingLibraryClientKey uniquely identifies context of netlify plumbing library client
	NetlifyAPIPlumbingLibraryClientKey ctxKey = 2 + iota
)

// KV identifiers
const (
	// NetlifyAuthTokenKVIdentifier is used to in suffix with userID to identify key in KV store
	NetlifyAuthTokenKVIdentifier            string = "_netlifyToken"
	NetlifyWebhookSubscriptionsKVIdentifier string = "_webhook"
)

// Netlify specific constants
const (
	// NetlifyAuthURL is auth url to make authentication to netlify
	NetlifyAuthURL string = "https://app.netlify.com/authorize"

	// NetlifyTokenURL is token url to get token from netlify
	NetlifyTokenURL string = "https://api.netlify.com/oauth/token"

	// NetlifyAPIHost is the base URL for making Netlify API request
	NetlifyAPIHost string = "api.netlify.com"

	// NetlifyAPIPath is path attached to baseURL for making Netlify API request
	NetlifyAPIPath string = "/api/v1"

	// NetlifyDateLayout is the date format returned by Netlify api for dates
	NetlifyDateLayout string = "2006-01-02T15:04:05.000Z"
)

// Netlify Build hook related
const (
	// MattermostNetlifyBuildHookTitle is the title of build hooks created by mattermost
	MattermostNetlifyBuildHookTitle string = "Mattermost-Netlify-Build-Hook"

	// MattermostNetlifyBuildHookMessage will be message of all build hook deploys from mattermost
	MattermostNetlifyBuildHookMessage string = "triggered by Netlify Bot from Mattermost"
)

// SuccessfullyNetlifyConnectedMessage is posted when /connect command is executed and completed
const SuccessfullyNetlifyConnectedMessage string = "#### Mattermost Netlify Plugin is now connected\n" +
	"You've successfully connected your Netlify account on Mattermost. To see more details about the account you can run `/netlify me`. For any other help run `/netlify help`\n\n" +
	"##### Notifications\n" +
	"You can run `/netlify subscribe` on any channel where you would like to receive build notifications for your Netlify site. To view all subscriptions for notification run `/netlify subscriptions`\n" +
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

	MarkdownSubscriptionTableHeader string = `
| Site | URL | Status |
|------|:---:|--------|`
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

// Netlify Notification Hook events types
const (
	NetlifyEventSubmissionCreated       string = "submission_created"
	NetlifyEventSplitTestActivated      string = "split_test_activated"
	NetlifyEventSplitTestDeactivated    string = "split_test_deactivated"
	NetlifyEventSplitTestModified       string = "split_test_modified"
	NetlifyEventLiveSessionConnected    string = "live_session_connected"
	NetlifyEventliveSessionDisconnected string = "live_session_disconnected"

	// NetlifyEventDeployBuilding is emitted when deploy is started
	NetlifyEventDeployBuilding string = "deploy_building"

	// NetlifyEventDeployCreated is emitted when deploy is successfull
	NetlifyEventDeployCreated string = "deploy_created"

	// NetlifyEventDeployFailed is emitted when deploy is failed
	NetlifyEventDeployFailed string = "deploy_failed"

	NetlifyEventDeployLocked          string = "deploy_locked"
	NetlifyEventDeployUnlocked        string = "deploy_unlocked"
	NetlifyEventDeployRequestPending  string = "deploy_request_pending"
	NetlifyEventDeployRequestAccepted string = "deploy_request_accepted"
	NetlifyEventDeployRequestRejected string = "deploy_request_rejected"
)

// Information of state inside of incoming webhook
const (
	NetlifyEventStateDeployCreated  string = "ready"
	NetlifyEventStateDeployBuilding string = "building"
	NetlifyEventStateDeployFailed   string = "error"
)

// Types of Netlify Hooks
const (
	NetlifyHookTypeSlack string = "slack"
	NetlifyHookTypeURL   string = "url"
	NetlifyHookTypeEmail string = "email"
)

// Header information inside of incoming webhook
const (
	NetlifyEventTypeHeader string = "X-Netlify-Event"
	NetlifyJWSHeader       string = "X-Webhook-Signature"
)

// HelpPost is string with all commands description
const HelpPost string = `
* /netlify **connect** - Connect your Mattermost account to your Netlify account. For any of the below commands, this command should be run first.
* /netlify **disconnect** - Disconnect your Mattermost account from your Netlify account All notifications are also unsubscribed from all channels.
* /netlify **list** - It tabulates all the sites information of Netlify account. It lists name, url, custom domain, repository, deployed branch, managed by team, last updated of the site.
* /netlify **list id** - This is usually a precursor command which you will be using to obtain site ids of you netlify hosted sites. It tabulates your sites along with its ids.
* /netlify **deploy** - Triggers a rebuild or build for your Netlify site.
* /netlify **rollback** - Facilitate to quick rollback to a previous stable state of your Netlify site.
* /netlify **subscribe** - Subscribes the channel to receive build notifications from your Netlify site(s).
* /netlify **unsubscribe** - Unsubscribes the channel from build notifications from all of your Netlify site(s).
* /netlify **subscriptions** - Lists out all your Netlify site(s) subscribed with the channel.
* /netlify **site** - Shows in-depth information of your Netlify site.
* /netlify **me** - This commands show revelant information of the Netlify account connected to Mattermost.
* /netlify **help** - Shows help with plugin commands and features.
`

// Ref : https://github.com/mattermost/mattermost-server/blob/v5.20.1/model/channel.go
const (
	MattermostChannelPublic  string = "O"
	MattermostChannelPrivate string = "P"
	MattermostChannelDM      string = "D"
	MattermostChannelGroup   string = "G"
)

// MattermostGetSitePost is post template used for site command
const MattermostGetSitePost string = `#### Netlify site information for %v

##### Basic details
*Name* : **%v** - %v
*URL* : %v
*Created at* : %v
*Published at* : %v
*Managed by* : %v

##### Domain details
*Custom domain* : %v
*Domain aliases* : %v
*Netlify managed DNS* : %v
*Enabled SSL* : %v
*Forced enabled SSL* : %v

##### Repository details
*Repository* : %v
*Repository public status* : %v
*Private logs* : %v

##### Enhancements (%v)
*Stylesheets*
*Bundle CSS* : %v
*Minify CSS* : %v

*Javascript*
*Bundle JS* : %v
*Minify JS* : %v

*Pretty URL* : %v
*Optimize images* : %v

**[Manage site at Netlify App](%v)**`
