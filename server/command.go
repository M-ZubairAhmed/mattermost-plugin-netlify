package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// Custom slash commands to setup
func getCommand() *model.Command {
	return &model.Command{
		Trigger:          "netlify",
		AutoComplete:     true,
		AutoCompleteHint: "[command]",
		AutoCompleteDesc: "Available commands: connect, disconnect",
		DisplayName:      "Netlify",
		Description:      "Integration with Netlify",
	}
}

// ExecuteCommand executes the commands registered on getCommand() via RegisterCommand hook
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	// Split the entered command based on white space
	arguments := strings.Fields(args.Command)

	// Eg. "netlify" in command "/netlify"
	baseCommand := arguments[0]

	// Eg "connect" in command "/netlify connect"
	action := ""
	if len(arguments) > 1 {
		action = arguments[1]
	}

	// Reject any command not prefixed netlify
	if baseCommand != "netlify" {
		return &model.CommandResponse{}, nil
	}

	switch action {
	case "connect":
		p.wrapperSendEphemeralPost(args, "Ok connect exec")
		return &model.CommandResponse{}, nil
	case "help":
		p.wrapperSendEphemeralPost(args, "help")
		return &model.CommandResponse{}, nil
	case "":
		p.wrapperSendEphemeralPost(args, "help fallback")
		return &model.CommandResponse{}, nil
	default:
		p.wrapperSendEphemeralPost(args, fmt.Sprintf("Unknown action %v", action))
		return &model.CommandResponse{}, nil
	}
}

func (p *Plugin) wrapperSendEphemeralPost(args *model.CommandArgs, text string) {
	post := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: args.ChannelId,
		Message:   text,
	}
	_ = p.API.SendEphemeralPost(args.UserId, post)
}
