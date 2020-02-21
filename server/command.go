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
		DisplayName:      "Netlify",
		Description:      "Integration with Netlify",
		AutoComplete:     true,
		AutoCompleteDesc: "Available commands: connect, help",
		AutoCompleteHint: "[command]",
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
	if baseCommand != "/netlify" {
		return &model.CommandResponse{}, nil
	}

	// /connect slash command to connect to Netlify
	if action == "connect" {
		// Check if SiteURL is defined in the app
		siteURL := p.API.GetConfig().ServiceSettings.SiteURL
		if siteURL == nil {
			p.sendBotEphemeralPostWithMessage(args, "Error! Site URL is not defined in the App")
			return &model.CommandResponse{}, nil
		}

		// Send an ephemeral post with the link to connect netlify
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf("[Click here to link your Netlify account.](%s/plugins/netlify/auth/connect)", *siteURL))
		return &model.CommandResponse{}, nil
	}

	// Before executing any of below commands check if user account is connected
	// TODO

	// /help slash command to list all netlify commands
	if action == "help" || action == "" {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf("Help"))
		return &model.CommandResponse{}, nil

	}

	// Unknown slash command if no action matches
	p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf("Unknown action %v, to see list of commands type `/netlify help`", action))
	return &model.CommandResponse{}, nil
}
