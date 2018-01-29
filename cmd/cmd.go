package cmd

import "github.com/spf13/cobra"

var (
	Port         string
	SlackToken   string
	SlackChannel string
	OutDir       string
	CompanyName  string
	WebsiteURL   string
	LogoURL      string
	TemplatesDir string

	Cmd = &cobra.Command{
		Use:   "contactification",
		Short: "contactification is a simple tool to send contact informations to slack",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
)

func init() {
	Cmd.PersistentFlags().StringVar(&Port, "port", "1789", "port to start the http server")
	Cmd.PersistentFlags().StringVar(&OutDir, "outDir", "./out", "output directory for new contacts request")
	Cmd.PersistentFlags().StringVar(&SlackChannel, "slackChannel", "", "slack channel in which to send the notifications")
	Cmd.PersistentFlags().StringVar(&SlackToken, "slackToken", "", "slack token for authentication")
	Cmd.PersistentFlags().StringVar(&CompanyName, "companyName", "", "company name to use with the slack bot")
	Cmd.PersistentFlags().StringVar(&WebsiteURL, "websiteURL", "", "website where the form is used")
	Cmd.PersistentFlags().StringVar(&LogoURL, "logoURL", "", "logo URL")
	Cmd.PersistentFlags().StringVar(&TemplatesDir, "templatesDir", "./templates", "template dir for the messages to display in slack")

	Cmd.MarkPersistentFlagRequired("slackChannel")
	Cmd.MarkPersistentFlagRequired("slackToken")
	Cmd.MarkPersistentFlagRequired("companyName")
	Cmd.MarkPersistentFlagRequired("websiteURL")
	Cmd.MarkPersistentFlagRequired("logoURL")
}
