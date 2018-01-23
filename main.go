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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
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

	errMethodNotAllowed   = errors.New("method not allowed")
	errPageNotFound       = errors.New("page not found")
	errInvalidContentType = errors.New("invalid content-type")

	cmd = &cobra.Command{
		Use:   "contactification",
		Short: "contactification is a simple tool to send contact informations to slack",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	slackClient *slack.Client
)

func init() {
	cmd.PersistentFlags().StringVar(&port, "port", "1789", "port to start the http server")
	cmd.PersistentFlags().StringVar(&outDir, "outDir", ".", "output directory for new contacts request")
	cmd.PersistentFlags().StringVar(&slackChannel, "slackChannel", "", "slack channel in which to send the notifications")
	cmd.PersistentFlags().StringVar(&slackToken, "slackToken", "", "slack token for authentication")

	cmd.MarkPersistentFlagRequired("slackChannel")
	cmd.MarkPersistentFlagRequired("slackToken")
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

func sendSlackNotification() {

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

	if ok, err := dirExists(); !ok {
		fmt.Printf("%v\n", err)
		return
	}

	if err := initSlack(); err != nil {
		fmt.Printf("%v\n", err)
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
	err := srv.Close()
	if err != nil {
		fmt.Printf("Error closing server: %s\n", err.Error())
	}
	fmt.Printf("Goodbye !\n")
}
