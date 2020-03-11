package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/go-openapi/runtime"
	openapiClient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/netlify/open-api/go/plumbing"
	"golang.org/x/oauth2"
)

// sendMessageFromBot can create a regular or ephemeral post on Channel or on DM from BOT.
// 1. "For DM Reg post" : [0]channelID, [X]userID, [0]isEphemeralPost.
// 2. "For DM Eph post" : [0]channelID, [X]userID, [X]isEphemeralPost.
// 3. "For Ch Reg post" : [X]channelID, [0]userID, [0]isEphemeralPost.
// 4. "For Ch Eph post" : [X]channelID, [X]userID, [X]isEphemeralPost.
func (p *Plugin) sendMessageFromBot(_channelID string, userID string, isEphemeralPost bool, message string) error {
	var channelID string = _channelID

	// If its nil then get the DM channel of bot and user
	if len(channelID) == 0 {
		if len(userID) == 0 {
			return errors.New("User and Channel ID both are undefined")
		}

		// Get the Bot Direct Message channel
		directChannel, err := p.API.GetDirectChannel(userID, p.BotUserID)
		if err != nil {
			return err
		}

		channelID = directChannel.Id
	}

	// Construct the Post message
	post := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: channelID,
		Message:   message,
	}

	if isEphemeralPost == true {
		p.API.SendEphemeralPost(userID, post)
		return nil
	}

	p.API.CreatePost(post)
	return nil

}

// setNetlifyUserAccessTokenToStore : Stores the access token along with userID inside of KV store
func (p *Plugin) setNetlifyUserAccessTokenToStore(token *oauth2.Token, userID string) error {
	// Convert the token to KV supported byte format
	accessToken := []byte(token.AccessToken)

	// Unique identifier
	accessTokenIdentifier := userID + NetlifyAuthTokenKVIdentifier

	// Store the accesstoken into KV store with a unique identifier i.e userid_netlifyToken
	// TODO : store encrypted version of Access Token
	// TODO : store complete *oauth2.Token strut
	err := p.API.KVSet(accessTokenIdentifier, accessToken)
	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) getNetlifyUserAccessTokenFromStore(userID string) (string, error) {
	// Unique identifier
	accessTokenIdentifier := userID + NetlifyAuthTokenKVIdentifier

	// Get the token from the KV store
	accessTokenInBytes, err := p.API.KVGet(accessTokenIdentifier)
	if err != nil || accessTokenInBytes == nil {
		return "", err
	}

	// TODO make use of ReuseTokenSource to automatically get new token when they expires
	// https://pkg.go.dev/golang.org/x/oauth2?tab=doc#ReuseTokenSource
	accessToken := string(accessTokenInBytes)

	return accessToken, nil
}

func (p *Plugin) getNetlifyClientCredentials(userID string) (runtime.ClientAuthInfoWriterFunc, error) {
	// Get access token from KV store
	accessToken, err := p.getNetlifyUserAccessTokenFromStore(userID)
	if err != nil || len(accessToken) == 0 {
		return nil, err
	}
	// Add OpenAPI runtime credentials
	openAPICredentials := runtime.ClientAuthInfoWriterFunc(
		func(r runtime.ClientRequest, _ strfmt.Registry) error {
			r.SetHeaderParam("User-Agent", "test")
			r.SetHeaderParam("Authorization", "Bearer "+accessToken)
			return nil
		})

	return openAPICredentials, nil
}

func (p *Plugin) getHTTPClient() *http.Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   -1,
			DisableKeepAlives:     true}}

	return httpClient
}

func (p *Plugin) getNetlifyClient() (*plumbing.Netlify, context.Context) {
	// Create OpenAPI transport
	transport := openapiClient.NewWithClient(NetlifyAPIHost, NetlifyAPIPath, plumbing.DefaultSchemes, p.getHTTPClient())
	transport.SetDebug(true)

	// Create Netlify client by adding the transport to it
	client := plumbing.New(transport, strfmt.Default)

	// Create an empty context
	ctx := context.Background()

	// Add client to that context
	ctx = context.WithValue(ctx, NetlifyAPIPorcelainLibraryClientKey, client)

	return client, ctx
}

// truncateString trims the given string to specified length.
func truncateString(s string, i int) string {
	runes := []rune(s)
	if len(runes) > i {
		return string(runes[:i])
	}
	return s + "..."
}

func (p *Plugin) isCommandRunFromValidChannel(channelID string) bool {
	channel, err := p.API.GetChannel(channelID)
	if err != nil {
		return false
	}

	if channel.Type == MattermostChannelDM || channel.Type == MattermostChannelGroup {
		return false
	}

	return true
}
