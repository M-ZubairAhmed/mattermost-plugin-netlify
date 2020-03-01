package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	netlifyModels "github.com/netlify/open-api/go/models"
	netlifyPlumbingModels "github.com/netlify/open-api/go/plumbing/operations"
)

// NetlifyWebhookEvent is the struct that properties similar to what a webhook return
type NetlifyWebhookEvent struct {
	Name         string `json:"name"`
	SiteID       string `json:"site_id"`
	BuildID      string `json:"build_id"`
	AdminURL     string `json:"admin_url"`
	State        string `json:"state"`
	ErrorMessage string `json:"error_message"`
	Branch       string `json:"branch"`
	DeploySSLURL string `json:"deploy_ssl_url"`
}

func (p *Plugin) handleWebhooks(w http.ResponseWriter, r *http.Request) {
	// If the body isn't of type json, then reject
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "Corrupt incoming webhook, Content types don't match", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Corrupt incoming webhook, Body data couldn't be read", http.StatusBadRequest)
		return
	}

	webhookEventData := NetlifyWebhookEvent{}

	err = json.Unmarshal(body, &webhookEventData)
	if err != nil {
		http.Error(w, "Corrupt incoming webhook, Cannot unmarshal input json", http.StatusBadRequest)
		return
	}

	// Construct the build log
	buildLogURL := fmt.Sprintf("%v/deploys/%v", webhookEventData.AdminURL, webhookEventData.BuildID)

	subscribedChannels, err := p.getWebhookSubscriptionForSite(webhookEventData.SiteID)

	if len(subscribedChannels) == 0 {
		http.Error(w, "No channels subscribed to the site", http.StatusNotFound)
		return
	}

	eventType := r.Header.Get(NetlifyEventTypeHeader)
	switch eventType {
	case NetlifyEventDeployBuilding:
		messageAttachment := &model.SlackAttachment{
			Fallback:  fmt.Sprintf("There is a new deploy in process for %v", webhookEventData.Name),
			Color:     "#c2a344",
			Pretext:   fmt.Sprintf(":flight_departure: There is a new deploy in process for **%v**", webhookEventData.Name),
			Title:     "Visit the build log",
			TitleLink: buildLogURL,
			Footer:    fmt.Sprintf("Using git %v branch", webhookEventData.Branch),
		}

		for _, channelID := range subscribedChannels {
			p.API.CreatePost(&model.Post{
				UserId:    p.BotUserID,
				ChannelId: channelID,
				Props: map[string]interface{}{
					"attachments": []*model.SlackAttachment{messageAttachment},
				},
			})
		}

		return
	case NetlifyEventDeployCreated:
		messageAttachment := &model.SlackAttachment{
			Fallback:  fmt.Sprintf("Successful deploy of %v", webhookEventData.Name),
			Color:     "#3ab259",
			Pretext:   fmt.Sprintf(":rocket: Successful deploy of **%v**", webhookEventData.Name),
			Title:     "Visit the changes live",
			TitleLink: webhookEventData.DeploySSLURL,
			Text:      fmt.Sprintf("Or check out the [build log](%v)", buildLogURL),
			Footer:    fmt.Sprintf("Using git %v branch", webhookEventData.Branch),
		}
		for _, channelID := range subscribedChannels {
			p.API.CreatePost(&model.Post{
				UserId:    p.BotUserID,
				ChannelId: channelID,
				Props: map[string]interface{}{
					"attachments": []*model.SlackAttachment{messageAttachment},
				},
			})
		}
		return
	case NetlifyEventDeployFailed:
		messageAttachment := &model.SlackAttachment{
			Fallback:  fmt.Sprintf("Something went wrong deploying %v", webhookEventData.Name),
			Color:     "#b2593a",
			Pretext:   fmt.Sprintf(":fire: Something went wrong deploying **%v**", webhookEventData.Name),
			Title:     "Visit the build log",
			TitleLink: buildLogURL,
			Text:      fmt.Sprintf("The last message we got from the build was `%v`", webhookEventData.ErrorMessage),
			Footer:    fmt.Sprintf("Using git %v branch", webhookEventData.Branch),
		}
		for _, channelID := range subscribedChannels {
			p.API.CreatePost(&model.Post{
				UserId:    p.BotUserID,
				ChannelId: channelID,
				Props: map[string]interface{}{
					"attachments": []*model.SlackAttachment{messageAttachment},
				},
			})
		}
		return
	default:
		http.Error(w, "Incoming webhook of unknown type", http.StatusBadRequest)
		// If it doesn't contain valid event, then reject
		return
	}
}

func (p *Plugin) handleSiteSelectionForSubscribeCommand(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON
	intergrationResponseFromCommand := model.PostActionIntegrationRequestFromJson(r.Body)

	originalPostID := intergrationResponseFromCommand.PostId
	channelIDToSubscribe := intergrationResponseFromCommand.ChannelId
	channelNameToSubscribe := intergrationResponseFromCommand.ChannelName
	userID := intergrationResponseFromCommand.UserId
	actionSecretPassed := intergrationResponseFromCommand.Context["actionSecret"].(string)
	actionSecret := p.getConfiguration().EncryptionKey
	selectedOption := intergrationResponseFromCommand.Context["selected_option"].(string)
	selectedOptionsValue := strings.Fields(selectedOption)
	siteIDToSubscribe := selectedOptionsValue[0]
	siteNameToSubscribe := selectedOptionsValue[1]

	// Check if this was passed within Mattermost
	authUserID := r.Header.Get("Mattermost-User-ID")
	if authUserID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		p.API.DeleteEphemeralPost(userID, originalPostID)
		return
	}

	// If action was not initiated from within MM
	if actionSecret != actionSecretPassed {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelIDToSubscribe,
			Message: fmt.Sprintf(
				":exclamation: Authentication failed\n"),
		})
		p.API.DeleteEphemeralPost(userID, originalPostID)
		return
	}

	// Check if any selected option is empty
	if len(selectedOptionsValue) == 0 || len(siteIDToSubscribe) == 0 || len(siteNameToSubscribe) == 0 {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelIDToSubscribe,
			Message: fmt.Sprintf(
				":exclamation: One of more values while selecting from dropdown were empty"),
		})
		p.API.DeleteEphemeralPost(userID, originalPostID)
		return
	}

	// Check if SiteURL is defined in the app
	siteURL := p.API.GetConfig().ServiceSettings.SiteURL
	if siteURL == nil {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelIDToSubscribe,
			Message: fmt.Sprintf(
				":exclamation: Error! Site URL is not defined in the App\n"),
		})
		p.API.DeleteEphemeralPost(userID, originalPostID)
		return
	}

	// Get the netlify client
	netlifyClient, ctx := p.getNetlifyClient()
	netlifyClientCredentials, err := p.getNetlifyClientCredentials(userID)
	if err != nil {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelIDToSubscribe,
			Message: fmt.Sprintf(
				":exclamation: Authentication failed\n"+
					"*Error : %v*", err.Error()),
		})
		p.API.DeleteEphemeralPost(userID, originalPostID)
		return
	}

	// Construct the same dropdown to update the original dropdown
	subscribeCommandDropdown := &model.PostAction{
		Type:     model.POST_ACTION_TYPE_SELECT,
		Name:     siteNameToSubscribe,
		Disabled: true,
		Options:  []*model.PostActionOptions{},
	}

	subscribeCommandAttachment := &model.SlackAttachment{
		Pretext: "Subscribe to Netlify notifications",
		Title:   "Select a site you want to subscribe build notifications for",
		Text:    "Selecting a site will subscribe current channel for build start, success and fail notifications.\n",
		Actions: []*model.PostAction{subscribeCommandDropdown},
		Footer:  "If however you don't wish to subscribe, hit the (x) cross icon on the right to dismiss this message",
	}

	subscribeCommandPost := &model.Post{
		Id:        originalPostID,
		UserId:    p.BotUserID,
		ChannelId: channelIDToSubscribe,
		Props: map[string]interface{}{
			"attachments": []*model.SlackAttachment{subscribeCommandAttachment},
		},
	}

	// Present the user with the site dropdown updated
	p.API.UpdateEphemeralPost(userID, subscribeCommandPost)

	p.sendMessageFromBot(channelIDToSubscribe, "", false,
		fmt.Sprintf(":hourglass: Hang on while subscribring is in progress for **%v** channel with **%v** build notifications.", channelNameToSubscribe, siteNameToSubscribe),
	)

	// Get all the hooks with the site
	listHooksBySiteIDParams := &netlifyPlumbingModels.ListHooksBySiteIDParams{
		SiteID:  siteIDToSubscribe,
		Context: ctx,
	}

	listHooksBySiteIDResponse, err := netlifyClient.Operations.ListHooksBySiteID(listHooksBySiteIDParams, netlifyClientCredentials)
	if err != nil {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelIDToSubscribe,
			Message: fmt.Sprintf(
				":exclamation: Failed to get **%v** site subscriptions.\n"+
					"*Error : %v*", siteIDToSubscribe, err.Error()),
		})
		return
	}

	siteHooks := listHooksBySiteIDResponse.GetPayload()

	var mattermostSubscriptionWebhookURL string = fmt.Sprintf("%v/plugins/netlify/webhook", *siteURL)
	var isMMHookForBuildStartPresent bool = false
	var isMMHookForBuildCompletePresent bool = false
	var isMMHookForBuildFailPresent bool = false

	for _, siteHook := range siteHooks {
		typeOfNetlifyWebhook := siteHook.Type
		eventOfNetlifyWebhook := siteHook.Event

		// Hook type should be of URL type and its URL should be of that plugin webhook url
		if typeOfNetlifyWebhook == NetlifyHookTypeURL {
			// If type is of URL, then it contains url data
			urlOfNetlifyWebhook := siteHook.Data.(interface{}).(map[string]interface{})["url"].(string)
			if urlOfNetlifyWebhook == mattermostSubscriptionWebhookURL {
				if eventOfNetlifyWebhook == NetlifyEventDeployBuilding {
					isMMHookForBuildStartPresent = true
				} else if eventOfNetlifyWebhook == NetlifyEventDeployCreated {
					isMMHookForBuildCompletePresent = true
				} else if eventOfNetlifyWebhook == NetlifyEventDeployFailed {
					isMMHookForBuildFailPresent = true
				}
			}
		}
	}

	// Create hooks if not present
	if isMMHookForBuildStartPresent == false {
		hook := &netlifyModels.Hook{
			SiteID: siteIDToSubscribe,
			Type:   NetlifyHookTypeURL,
			Data: &map[string]interface{}{
				"url": mattermostSubscriptionWebhookURL,
			},
			Disabled: false,
			Event:    NetlifyEventDeployBuilding,
		}

		createHookBySiteIDParams := &netlifyPlumbingModels.CreateHookBySiteIDParams{
			Hook:    hook,
			SiteID:  siteIDToSubscribe,
			Context: ctx,
		}

		_, err := netlifyClient.Operations.CreateHookBySiteID(createHookBySiteIDParams, netlifyClientCredentials)
		if err != nil {
			p.API.SendEphemeralPost(userID, &model.Post{
				UserId:    p.BotUserID,
				ChannelId: channelIDToSubscribe,
				Message: fmt.Sprintf(
					":exclamation: Failed to create build notification of type `%v` for **%v** site.\n"+
						"*Error : %v*", NetlifyEventDeployBuilding, siteNameToSubscribe, err.Error()),
			})
			return
		}
		p.API.CreatePost(&model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelIDToSubscribe,
			Message: fmt.Sprintf(
				":fishing_pole_and_fish: Created a new webhook on Netlify of `%v` for **%v** site.\n",
				NetlifyEventDeployBuilding, siteNameToSubscribe),
		})

		isMMHookForBuildStartPresent = true
	}

	if isMMHookForBuildCompletePresent == false {
		hook := &netlifyModels.Hook{
			SiteID: siteIDToSubscribe,
			Type:   NetlifyHookTypeURL,
			Data: &map[string]interface{}{
				"url": mattermostSubscriptionWebhookURL,
			},
			Disabled: false,
			Event:    NetlifyEventDeployCreated,
		}

		createHookBySiteIDParams := &netlifyPlumbingModels.CreateHookBySiteIDParams{
			Hook:    hook,
			SiteID:  siteIDToSubscribe,
			Context: ctx,
		}

		_, err := netlifyClient.Operations.CreateHookBySiteID(createHookBySiteIDParams, netlifyClientCredentials)
		if err != nil {
			p.API.SendEphemeralPost(userID, &model.Post{
				UserId:    p.BotUserID,
				ChannelId: channelIDToSubscribe,
				Message: fmt.Sprintf(
					":exclamation: Failed to create build notification of type `%v` for **%v** site.\n"+
						"*Error : %v*", NetlifyEventDeployCreated, siteNameToSubscribe, err.Error()),
			})
			return
		}
		p.API.CreatePost(&model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelIDToSubscribe,
			Message: fmt.Sprintf(
				":fishing_pole_and_fish: Created a new webhook on Netlify of `%v` for **%v** site.\n",
				NetlifyEventDeployCreated, siteNameToSubscribe),
		})

		isMMHookForBuildCompletePresent = true
	}

	if isMMHookForBuildFailPresent == false {
		hook := &netlifyModels.Hook{
			SiteID: siteIDToSubscribe,
			Type:   NetlifyHookTypeURL,
			Data: &map[string]interface{}{
				"url": mattermostSubscriptionWebhookURL,
			},
			Disabled: false,
			Event:    NetlifyEventDeployFailed,
		}

		createHookBySiteIDParams := &netlifyPlumbingModels.CreateHookBySiteIDParams{
			Hook:    hook,
			SiteID:  siteIDToSubscribe,
			Context: ctx,
		}

		_, err := netlifyClient.Operations.CreateHookBySiteID(createHookBySiteIDParams, netlifyClientCredentials)
		if err != nil {
			p.API.SendEphemeralPost(userID, &model.Post{
				UserId:    p.BotUserID,
				ChannelId: channelIDToSubscribe,
				Message: fmt.Sprintf(
					":exclamation: Failed to create build notification of type `%v` for **%v** site.\n"+
						"*Error : %v*", NetlifyEventDeployFailed, siteNameToSubscribe, err.Error()),
			})
			return
		}
		p.API.CreatePost(&model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelIDToSubscribe,
			Message: fmt.Sprintf(
				":fishing_pole_and_fish: Created a new webhook on Netlify of `%v` for **%v** site.\n",
				NetlifyEventDeployFailed, siteNameToSubscribe),
		})

		isMMHookForBuildFailPresent = true
	}

	if isMMHookForBuildCompletePresent == true && isMMHookForBuildFailPresent == true && isMMHookForBuildStartPresent == true {
		err = p.setWebhookSubscriptionsForSite(channelIDToSubscribe, siteIDToSubscribe)
		if err != nil {
			p.API.SendEphemeralPost(userID, &model.Post{
				UserId:    p.BotUserID,
				ChannelId: channelIDToSubscribe,
				Message: fmt.Sprintf(
					":exclamation: Failed to subscribe build notification for **%v** site.\n"+
						"*Error : %v*", siteNameToSubscribe, err.Error()),
			})
			return
		}

		p.API.CreatePost(&model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelIDToSubscribe,
			Message: fmt.Sprintf(
				":star2:  Successfully subscribed **%v** for build notifications from **%v** site.\n",
				channelNameToSubscribe, siteNameToSubscribe),
		})
	}
	p.API.SendEphemeralPost(userID, &model.Post{
		UserId:    p.BotUserID,
		ChannelId: channelIDToSubscribe,
		Message: fmt.Sprintf(
			":grey_exclamation: Failed partially to subscribe build notification for **%v** site.\n"+
				"*Error : %v*", siteNameToSubscribe, err.Error()),
	})
	return
}

func removeDuplicatesFromSlice(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func (p *Plugin) setWebhookSubscriptionsForSite(channelID, siteID string) error {
	// Unique identifier
	webhookIdentifier := siteID + NetlifyWebhookSubscriptionsKVIdentifier

	channelsSubscribedTo, err := p.getWebhookSubscriptionForSite(siteID)
	if err != nil {
		return err
	}

	newChannelsSubscribedTo := append(channelsSubscribedTo, channelID)
	newChannelSubscriptions := removeDuplicatesFromSlice(newChannelsSubscribedTo)

	newChannelSubscriptionsString := strings.Join(newChannelSubscriptions, " ")
	newChannelSubscriptionsByte := []byte(newChannelSubscriptionsString)

	appErr := p.API.KVSet(webhookIdentifier, newChannelSubscriptionsByte)
	if appErr != nil {
		return appErr
	}

	return nil
}

func (p *Plugin) getWebhookSubscriptionForSite(siteID string) ([]string, error) {
	webhookIdentifier := siteID + NetlifyWebhookSubscriptionsKVIdentifier

	// Get the value from the store
	channelsSubscribedToBytes, err := p.API.KVGet(webhookIdentifier)
	if err != nil {
		return []string{}, err
	}

	// It returns nil if value is not found
	if channelsSubscribedToBytes == nil {
		return []string{}, nil
	}

	// If key is found then
	// Convert value to string
	channelsSubscribedToString := string(channelsSubscribedToBytes)

	// If it has no value
	if len(channelsSubscribedToString) == 0 {
		return []string{}, nil
	}

	// Split based on spaces
	channelsSubscribedTo := strings.Fields(channelsSubscribedToString)

	// If no values were seperated then
	if len(channelsSubscribedTo) == 0 {
		return []string{}, nil
	}

	return channelsSubscribedTo, nil
}
