package main

import (
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

var CertCmd = &cobra.Command{
	Use:   "cert",
	Short: "Manage certificate lifecycle",
	Long:  "Kubernetes-style certificate lifecycle operations: check certificate expiry and trigger renewal.",
}

var certStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check certificate expiry status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cert, _ := cmd.Flags().GetString("cert")
		cert = strings.TrimSpace(cert)

		path := "/api/admin/certificates/status"
		if cert != "" {
			path += "?cert=" + url.QueryEscape(cert)
		}

		return getAndPrint(path)
	},
}

var certRenewCmd = &cobra.Command{
	Use:   "renew",
	Short: "Trigger certificate renewal",
	RunE: func(cmd *cobra.Command, args []string) error {
		cert, _ := cmd.Flags().GetString("cert")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		payload := map[string]interface{}{
			"dry_run": dryRun,
		}
		if trimmed := strings.TrimSpace(cert); trimmed != "" {
			payload["cert"] = trimmed
		}

		return postAndPrint("/api/admin/certificates/renew", payload)
	},
}

func init() {
	certStatusCmd.Flags().String("cert", "", "Kubeadm cert name (e.g. apiserver, etcd-server, all)")
	certRenewCmd.Flags().String("cert", "", "Kubeadm cert name (e.g. apiserver, etcd-server, all)")
	certRenewCmd.Flags().Bool("dry-run", false, "Prepare renewal command without executing it")

	CertCmd.AddCommand(certStatusCmd, certRenewCmd)
}
