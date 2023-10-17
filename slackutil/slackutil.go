package slackutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"
	"time"

	"github.com/Threestup/aporosa/cmd"
	"github.com/slack-go/slack"
)

var (
	slackClient *slack.Client
)

// Init initialize the package
func Init() error {
	slackClient = slack.New(cmd.SlackToken)
	if slackClient == nil {
		return errors.New("unable to initialize slack client, please check your configuration")
	}
	return nil
}

// Notify sends a new notification to the Slack channel
func Notify(tpl *template.Template, values map[string]string) error {
	// Use the template to generate the message
	var msg bytes.Buffer
	err := tpl.Execute(&msg, values)
	if err != nil {
		return fmt.Errorf("unable to build template: %v", err)
	}

	attachment := slack.Attachment{
		AuthorIcon: cmd.LogoURL,
		AuthorName: cmd.CompanyName,
		Title:      fmt.Sprintf("New contact request for %s", cmd.CompanyName),
		TitleLink:  cmd.WebsiteURL,
		Footer:     "New contact request",
		Ts:         json.Number(fmt.Sprintf("%v", time.Now().Unix())),
		Text:       msg.String(),
		MarkdownIn: []string{"text"},
		ThumbURL:   cmd.LogoURL,
	}

	// Create a message option for attaching the attachment
	options := []slack.MsgOption{
		slack.MsgOptionText("", false),         // The text is empty here
		slack.MsgOptionAttachments(attachment), // Attach the attachment here
	}

	// Specify the channel where you want to send the message
	channelID := cmd.SlackChannel

	// Send the message
	_, _, err = slackClient.PostMessage(channelID, options...)

	if err != nil {
		return fmt.Errorf("unable to send slack message: %v", err)
	}

	return nil
}
