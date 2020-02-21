package main

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
)

// sendBotEphemeralPostWithMessage : Sends an ephemeral bot post to the channel from which slash command was executed
func (p *Plugin) sendBotEphemeralPostWithMessage(args *model.CommandArgs, text string) {
	post := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: args.ChannelId,
		Message:   text,
	}
	sentPost := p.API.SendEphemeralPost(args.UserId, post)
	fmt.Printf(sentPost.Message)
}
