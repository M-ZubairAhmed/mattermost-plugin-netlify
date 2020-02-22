package main

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"golang.org/x/oauth2"
)

// sendBotEphemeralPostWithMessage : Sends an ephemeral bot post to the channel from which slash command was executed
func (p *Plugin) sendBotEphemeralPostWithMessage(args *model.CommandArgs, text string) {
	post := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: args.ChannelId,
		Message:   text,
	}
	p.API.SendEphemeralPost(args.UserId, post)
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
