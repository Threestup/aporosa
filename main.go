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
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"text/template"
	"time"

	"github.com/Threestup/aporosa/cmd"
	"github.com/Threestup/aporosa/export"
	"github.com/Threestup/aporosa/slackutil"
	"github.com/Threestup/aporosa/templateutil"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/rs/cors"
)

const (
	basePath              = "/aporosa"
	healthCheckPath       = basePath + "/healthcheck"
	urlEncodedContentType = "application/x-www-form-urlencoded"
)

var (
	errMethodNotAllowed   = errors.New("method not allowed")
	errPageNotFound       = errors.New("page not found")
	errInvalidContentType = errors.New("invalid content-type")
)

type handler struct{}

func earlyExitWithError(rw http.ResponseWriter, r *http.Request, err error, status int) {
	fmt.Printf("from=\"%v\" error=\"%v\" ts=\"%v\"\n",
		r.RemoteAddr, err, time.Now().Format("Mon, 2 Jan 2006 15:04:05 MST"))
	rw.WriteHeader(status)
	fmt.Fprintf(rw, err.Error())
}

func (h handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		fmt.Printf("method=%v path=%v from=%v ts=%v time-taken=%v\n",
			r.Method, r.URL.Path, r.RemoteAddr, time.Now().Format("Mon, 2 Jan 2006 15:04:05 MST"), time.Since(start))
	}()

	if r.URL.Path == healthCheckPath {
		rw.WriteHeader(http.StatusOK)
		fmt.Fprintf(rw, "")
		return
	}

	// first check the request is well formated
	if r.Method != http.MethodPost {
		earlyExitWithError(rw, r, errMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	// try to match the path with one of the existings templates
	var tpl *template.Template
	for k, v := range templateutil.TemplatesMessages {
		if path.Clean(r.URL.Path) == path.Clean(path.Join(basePath, k)) {
			tpl = v
			break
		}
	}

	if tpl == nil {
		earlyExitWithError(rw, r, fmt.Errorf("%v:%v", errPageNotFound, r.URL.Path), http.StatusNotFound)
		return
	}

	// then extract form values

	if err := r.ParseForm(); err != nil {
		earlyExitWithError(rw, r, err, http.StatusInternalServerError)
		return
	}

	values := map[string]string{}
	for k, v := range r.PostForm {
		values[k] = v[0]
	}

	// then write them to a new_file
	if err := export.Save(tpl.Name(), values); err != nil {
		earlyExitWithError(rw, r, err, http.StatusInternalServerError)
		return
	}

	// then send to slack
	if err := slackutil.Notify(tpl, values); err != nil {
		earlyExitWithError(rw, r, err, http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	fmt.Fprintf(rw, "")
}

func dirExists(s string) (bool, error) {
	_, err := os.Stat(s)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, fmt.Errorf("%s: no such file or directory", s)
	}
	return true, err
}

func main() {
	if err := cmd.Cmd.Execute(); err != nil {
		return
	}

	if cmd.HelpFlag {
		os.Exit(0)
	}

	// check output dir exsits
	if ok, err := dirExists(cmd.OutDir); !ok {
		fmt.Printf("%v\n", err)
		return
	}

	// check output dir exsits
	if ok, err := dirExists(cmd.TemplatesDir); !ok {
		fmt.Printf("%v\n", err)
		return
	}

	// initialize slack client
	if err := slackutil.Init(); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	// generate messages template
	if err := templateutil.LoadFromDir(cmd.TemplatesDir); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	// init export package
	if err := export.Init(cmd.ExportMode); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	fmt.Printf("Available forms:\n")
	for k := range templateutil.TemplatesMessages {
		fmt.Printf("\tPOST %v\n", path.Join(basePath, k))
	}

	//1, nil), HelloHandler))
	srv := http.Server{
		Addr: ":" + cmd.Port,
		Handler: tollbooth.LimitHandler(
			tollbooth.NewLimiter(1, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Minute * 30}),
			cors.AllowAll().Handler(handler{})),
	}
	go func() {
		fmt.Printf("Server started on :%s\n", cmd.Port)
		srv.ListenAndServe()
	}()

	_ = <-sigc
	err := srv.Close()
	if err != nil {
		fmt.Printf("Error closing server: %s\n", err.Error())
	}
	fmt.Printf("Goodbye !\n")
}
