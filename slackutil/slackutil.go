package slackutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"
	"time"

	"github.com/Threestup/aporosa/cmd"
	"github.com/nlopes/slack"
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

// Notify send a new notification to the slack channel
func Notify(tpl *template.Template, values map[string]string) error {
	// use template to generate the message
	var msg bytes.Buffer
	err := tpl.Execute(&msg, values)
	if err != nil {
		return fmt.Errorf("unable to build template: %v", err)
	}

	att := slack.Attachment{
		AuthorIcon: cmd.LogoURL,
		AuthorName: cmd.CompanyName,
		Title:      fmt.Sprintf("New contact request for %s", cmd.CompanyName),
		TitleLink:  cmd.WebsiteURL,
		Footer:     "New contact request",
		Ts:         json.Number(fmt.Sprintf("%v", time.Now().Unix())),
		Text:       fmt.Sprintf(msg.String()),
		MarkdownIn: []string{"text"},
		ThumbURL:   cmd.LogoURL,
	}

	_, _, err = slackClient.PostMessage(
		cmd.SlackChannel,
		"",
		slack.PostMessageParameters{
			EscapeText: true,
			Username:   "NewContact",
			AsUser:     true,
			// IconURL:     "https://.slack.com/team/jeremy",
			Attachments: []slack.Attachment{att},
		},
	)

	if err != nil {
		return fmt.Errorf("unable to send slack message: %v", err)
	}

	return nil
}
