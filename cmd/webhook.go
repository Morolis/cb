package cmd

import (
	"fmt"
	"strings"

	"github.com/Morolis/cb/internal/api"
	"github.com/Morolis/cb/internal/config"
	"github.com/spf13/cobra"
)

var webhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Manage webhooks for event notifications",
	Long:  `Create, list, and manage webhooks that receive HTTP POST when snippets change.`,
}

var webhookBody string

var webhookAddCmd = &cobra.Command{
	Use:   "add <name> <url> <events>",
	Short: "Add a webhook",
	Long: `Add a webhook. Events: created, updated, deleted (comma-separated).
Example: cb webhook add myhook https://example.com/hook created,updated

Use --body to set a custom payload template. Write JSON with {{.Variable}} placeholders.
Leave --body empty to send the default JSON payload.

Available template variables:
  {{.Event}}              event type: created / updated / deleted
  {{.DateTime}}           ISO8601 timestamp
  {{.Snippet.ID}}         snippet UUID
  {{.Snippet.UserID}}     owner user ID
  {{.Snippet.Alias}}      alias name
  {{.Snippet.Description}} description
  {{.Snippet.Content}}    full content
  {{.Snippet.Encrypted}}  true / false
  {{.Snippet.Category}}   category
  {{.Snippet.Language}}   language hint
  {{.Snippet.ExpiresAt}}  expiry time (may be empty)
  {{.Snippet.CreatedAt}}  creation time
  {{.Snippet.UpdatedAt}}  last update time

Examples:
  cb webhook add slack https://hooks.slack.com/xxx created --body '{"text":"[{{.Event}}] {{.Snippet.Content}}"}'
  cb webhook add feishu https://open.feishu.cn/open-apis/bot/v2/hook/xxx created --body '{"msg_type":"text","content":{"text":"[{{.Event}}] {{.Snippet.Content}}"}'`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		client := api.NewClient(cfg)

		name := args[0]
		url := args[1]
		events := strings.Split(args[2], ",")

		resp, err := client.CreateWebhook(name, url, events, webhookBody)
		if err != nil {
			return fmt.Errorf("create webhook: %w", err)
		}

		fmt.Printf("Webhook created: %s (id: %s)\n", resp.Name, resp.ID)
		fmt.Printf("  URL: %s\n", resp.URL)
		fmt.Printf("  Events: %s\n", strings.Join(resp.Events, ", "))
		if resp.BodyTemplate != "" {
			fmt.Printf("  Template: %s\n", resp.BodyTemplate)
		}
		return nil
	},
}

var webhookListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all webhooks",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		client := api.NewClient(cfg)

		webhooks, err := client.ListWebhooks()
		if err != nil {
			return fmt.Errorf("list webhooks: %w", err)
		}

		if len(webhooks) == 0 {
			fmt.Println("No webhooks configured.")
			return nil
		}

		for _, w := range webhooks {
			status := "active"
			if !w.Active {
				status = "inactive"
			}
			fmt.Printf("  %s [%s]\n", w.Name, status)
			fmt.Printf("    ID:     %s\n", w.ID)
			fmt.Printf("    URL:    %s\n", w.URL)
			fmt.Printf("    Events: %s\n", strings.Join(w.Events, ", "))
			if w.BodyTemplate != "" {
				tmpl := w.BodyTemplate
				if len(tmpl) > 80 {
					tmpl = tmpl[:80] + "..."
				}
				fmt.Printf("    Template: %s\n", tmpl)
			}
			fmt.Println()
		}
		return nil
	},
}

var webhookRmCmd = &cobra.Command{
	Use:   "rm <id>",
	Short: "Delete a webhook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		client := api.NewClient(cfg)

		if err := client.DeleteWebhook(args[0]); err != nil {
			return fmt.Errorf("delete webhook: %w", err)
		}

		fmt.Printf("Webhook %s deleted.\n", args[0])
		return nil
	},
}

var webhookLogsCmd = &cobra.Command{
	Use:   "logs <id>",
	Short: "View webhook delivery logs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		client := api.NewClient(cfg)

		logs, err := client.ListWebhookLogs(args[0])
		if err != nil {
			return fmt.Errorf("list webhook logs: %w", err)
		}

		if len(logs) == 0 {
			fmt.Println("No delivery logs.")
			return nil
		}

		for _, l := range logs {
			status := "OK"
			if l.StatusCode == 0 {
				status = "FAIL"
			} else if l.StatusCode >= 400 {
				status = fmt.Sprintf("ERR %d", l.StatusCode)
			}
			fmt.Printf("  [%s] %s %s", l.CreatedAt, l.EventType, status)
			if l.Error != "" {
				fmt.Printf(" (%s)", l.Error)
			}
			fmt.Println()
		}
		return nil
	},
}

func init() {
	webhookAddCmd.Flags().StringVar(&webhookBody, "body", "", "custom payload template (Go text/template)")
	webhookCmd.AddCommand(webhookAddCmd)
	webhookCmd.AddCommand(webhookListCmd)
	webhookCmd.AddCommand(webhookRmCmd)
	webhookCmd.AddCommand(webhookLogsCmd)
	rootCmd.AddCommand(webhookCmd)
}
