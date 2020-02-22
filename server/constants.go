package main

// NetlifyAuthTokenKVIdentifier is used to in suffix with userID to identify key in KV store
const NetlifyAuthTokenKVIdentifier string = "_netlifyToken"

// NetlifyAuthURL is auth url to make authentication to netlify
const NetlifyAuthURL = "https://app.netlify.com/authorize"

// NetlifyTokenURL is token url to get token from netlify
const NetlifyTokenURL = "https://api.netlify.com/oauth/token"

// SuccessfullyNetlifyConnectedMessage is posted when /connect command is executed and completed
const SuccessfullyNetlifyConnectedMessage string = "#### Welcome to the Mattermost Netlify Plugin!\n" +
	"You've successfully connected your Mattermost account on Netlify.\n\n" +
	"##### Notifications\n" +
	"Write about how to enable notifications.\n" +
	"##### Slash Commands\n"