package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
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

	channelID := "qsumjxoun7yb7xbjyob7wsgjeh"
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

		p.API.CreatePost(&model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Props: map[string]interface{}{
				"attachments": []*model.SlackAttachment{messageAttachment},
			},
		})

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

		p.API.CreatePost(&model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Props: map[string]interface{}{
				"attachments": []*model.SlackAttachment{messageAttachment},
			},
		})
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

		p.API.CreatePost(&model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Props: map[string]interface{}{
				"attachments": []*model.SlackAttachment{messageAttachment},
			},
		})
		return
	default:
		http.Error(w, "Incoming webhook of unknown type", http.StatusBadRequest)
		// If it doesn't contain valid event, then reject
		return
	}
}
