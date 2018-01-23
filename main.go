// Copyright [2017] [threestup]
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"text/template"
	"time"

	"github.com/nlopes/slack"
	"github.com/spf13/cobra"
)

const (
	contactNotificationPath = "/contact-notification"
	urlEncodedContentType   = "application/x-www-form-urlencoded"
)

var (
	port         string
	slackToken   string
	slackChannel string
	outDir       string
	companyName  string
	websiteURL   string
	logoURL      string
	message      string

	errMethodNotAllowed   = errors.New("method not allowed")
	errPageNotFound       = errors.New("page not found")
	errInvalidContentType = errors.New("invalid content-type")

	cmd = &cobra.Command{
		Use:   "contactification",
		Short: "contactification is a simple tool to send contact informations to slack",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	slackClient *slack.Client

	messageTemplate = template.New("Slack message")
)

func init() {
	cmd.PersistentFlags().StringVar(&port, "port", "1789", "port to start the http server")
	cmd.PersistentFlags().StringVar(&outDir, "outDir", ".", "output directory for new contacts request")
	cmd.PersistentFlags().StringVar(&slackChannel, "slackChannel", "", "slack channel in which to send the notifications")
	cmd.PersistentFlags().StringVar(&slackToken, "slackToken", "", "slack token for authentication")
	cmd.PersistentFlags().StringVar(&companyName, "companyName", "", "company name to use with the slack bot")
	cmd.PersistentFlags().StringVar(&websiteURL, "websiteURL", "", "website where the form is used")
	cmd.PersistentFlags().StringVar(&logoURL, "logoURL", "", "logo URL")
	cmd.PersistentFlags().StringVar(&message, "message", "", "template file for the message to display in slack")

	cmd.MarkPersistentFlagRequired("slackChannel")
	cmd.MarkPersistentFlagRequired("slackToken")
	cmd.MarkPersistentFlagRequired("companyName")
	cmd.MarkPersistentFlagRequired("websiteURL")
	cmd.MarkPersistentFlagRequired("logoURL")
	cmd.MarkPersistentFlagRequired("message")
}

type handler struct{}

func (h handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// first check the request is well formated

	if r.Method != http.MethodPost {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(rw, errMethodNotAllowed.Error())
		return
	}

	if r.URL.Path != contactNotificationPath {
		rw.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rw, errPageNotFound.Error())
		return
	}

	// then extract form values

	if err := r.ParseForm(); err != nil {
		fmt.Printf("%v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, err.Error())
	}

	values := map[string]string{}
	for k, v := range r.PostForm {
		values[k] = v[0]
	}

	// then write them to a new_file
	if err := saveContact(values); err != nil {
		fmt.Printf("%v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, err.Error())
	}

	// then send to slack
	if err := sendSlackNotification(values); err != nil {
		fmt.Printf("%v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, err.Error())
	}

	rw.WriteHeader(http.StatusOK)
	fmt.Fprintf(rw, "all good")
}

func saveContact(values map[string]string) error {
	fmt.Printf("new contact infos: %v\n", values)

	// make json
	b, err := json.Marshal(values)
	if err != nil {
		return fmt.Errorf("unable to marshal contact infos: %v", err)
	}

	// writing file
	now := fmt.Sprintf("%v", time.Now().Unix())
	filePath := path.Join(outDir, now+".contact.json")
	err = ioutil.WriteFile(filePath, b, 0644)
	if err != nil {
		return fmt.Errorf("unable to write file: %v", err)
	}

	return nil
}

func dirExists() (bool, error) {
	_, err := os.Stat(outDir)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, fmt.Errorf("%s: no such file or directory", outDir)
	}
	return true, err
}

func sendSlackNotification(values map[string]string) error {
	// use template to generate the message
	var msg bytes.Buffer
	err := messageTemplate.Execute(&msg, values)
	if err != nil {
		return fmt.Errorf("unable to build template: %v", err)
	}

	att := slack.Attachment{
		AuthorIcon: logoURL,
		AuthorName: companyName,
		Title:      fmt.Sprintf("New contact request for %s", companyName),
		TitleLink:  websiteURL,
		Footer:     "New contact request",
		Ts:         json.Number(fmt.Sprintf("%v", time.Now().Unix())),
		Text:       fmt.Sprintf(msg.String()),
		MarkdownIn: []string{"text"},
		ThumbURL:   logoURL,
	}

	_, _, err = slackClient.PostMessage(
		slackChannel,
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

func initSlack() error {
	slackClient = slack.New(slackToken)
	if slackClient == nil {
		return errors.New("unable to initialize slack client, please check your configuration")
	}
	return nil
}

func main() {
	if err := cmd.Execute(); err != nil {
		return
	}

	// check output dir exsits
	if ok, err := dirExists(); !ok {
		fmt.Printf("%v\n", err)
		return
	}

	// initialize slack client
	if err := initSlack(); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	// generate messages template
	messageBytes, err := ioutil.ReadFile(message)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	messageTemplate, err = messageTemplate.Parse(string(messageBytes))
	if err != nil {
		fmt.Printf("unable to parse message template: %v", err)
		return
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	srv := http.Server{Addr: ":" + port, Handler: handler{}}

	go func() {
		fmt.Printf("Server started on :%s\n", port)
		srv.ListenAndServe()
	}()

	_ = <-sigc
	err = srv.Close()
	if err != nil {
		fmt.Printf("Error closing server: %s\n", err.Error())
	}
	fmt.Printf("Goodbye !\n")
}
