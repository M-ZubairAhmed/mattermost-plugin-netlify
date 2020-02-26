package main

import (
	"context"
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

// sendBotEphemeralPostWithMessage : Sends an ephemeral bot post to the channel from which slash command was executed
func (p *Plugin) sendBotEphemeralPostWithMessage(args *model.CommandArgs, text string) {
	post := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: args.ChannelId,
		Message:   text,
	}
	p.API.SendEphemeralPost(args.UserId, post)
}

func (p *Plugin) sendBotEphemeralPostWithMessageInChannel(channelID string, userID string, text string) {
	post := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: channelID,
		Message:   text,
	}
	p.API.SendEphemeralPost(userID, post)
}

func (p *Plugin) sendBotPostOnChannel(args *model.CommandArgs, text string) *model.AppError {
	// Construct the Post message
	post := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: args.ChannelId,
		Message:   text,
	}

	// Send the Post
	_, err := p.API.CreatePost(post)
	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) createBotPost(channelID string, text string) *model.AppError {
	// Construct the Post message
	post := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: channelID,
		Message:   text,
	}

	// Send the Post
	_, err := p.API.CreatePost(post)
	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) sendBotPostOnDM(userID string, message string) *model.AppError {
	// Get the Bot Direct Message channel
	directChannel, err := p.API.GetDirectChannel(userID, p.BotUserID)
	if err != nil {
		return err
	}

	// Construct the Post message
	post := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: directChannel.Id,
		Message:   message,
	}

	// Send the Post
	_, err = p.API.CreatePost(post)
	if err != nil {
		return err
	}

	return nil
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
