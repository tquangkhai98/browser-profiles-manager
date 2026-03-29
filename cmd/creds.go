package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tquangkhai98/browser-profiles-manager/internal/credential"
	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
)

var credsJSON bool

var credsCmd = &cobra.Command{
	Use:   "creds <name>",
	Short: "Inspect credentials in a browser profile",
	Long:  `Show which sites have cookies and/or saved logins in a profile. This only reads domain names and counts — values are never decrypted.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		p, err := profile.Get(name)
		if err != nil {
			return err
		}

		result, err := credential.Inspect(p.DataDir, p.Name)
		if err != nil {
			return err
		}

		if credsJSON {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if len(result.Sites) == 0 {
			fmt.Printf("No credentials found in profile %q\n", name)
			fmt.Println("Use the profile first with: bpm use " + name)
			return nil
		}

		fmt.Printf("Credentials in profile %q:\n\n", name)
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "DOMAIN\tCOOKIES\tLOGINS")
		fmt.Fprintln(w, "------\t-------\t------")
		for _, s := range result.Sites {
			fmt.Fprintf(w, "%s\t%d\t%d\n", s.Domain, s.CookieCount, s.LoginCount)
		}
		w.Flush()

		fmt.Printf("\nTotal: %d cookies, %d logins\n", result.TotalCookies, result.TotalLogins)
		return nil
	},
}

func init() {
	credsCmd.Flags().BoolVar(&credsJSON, "json", false, "Output in JSON format")
	rootCmd.AddCommand(credsCmd)
}
