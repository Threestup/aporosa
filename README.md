aporosa
=======
[![Build Status](https://travis-ci.org/Threestup/aporosa.svg?branch=master)](https://travis-ci.org/Threestup/aporosa) [![Go Report Card](https://goreportcard.com/badge/github.com/threestup/aporosa)](https://goreportcard.com/report/github.com/threestup/aporosa)

API service to handle simples web form and create slack notifications from them.

### Templates
Templates are used to format the slack notification. It's basically using the Go `text/template`.
You can find templates examples in the /templates folder.

They are also used to create the API endpoints based on the names of the templates files, e.g if your template is called `contact-us.tpl`, this will generate the following endpoint:
> POST /aporosa/contact-us

### Usage example with Docker
```bash
docker pull threestup/aporosa
docker run --rm -d -p 8080:8080 \
	-v "`pwd`/out:/out" \ # json forms output folder
	-v "`pwd`/templates:/templates" \ # templates input folder
	threestup/aporosa \
	--slackToken="YOUR_SLACK_TOKEN" \
	--slackChannel="SLACK_CHANNEL" \
	--companyName="Threestup" \
	--logoURL="https://threestup.com/ts-logomark-64.png" \
	--websiteURL="https://threestup.com" \
	--port=8080
  --exportMode=JSON
```

Also you can clone the git repository, and use the docker-compose configuration to run the service, you will just need toupdate the .env file in order to set the parameter you to pass to aporosa.

### How to build locally
Just clone the repository, then go to the root of the repository and run:
> go build