package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/claudioed/deployment-tail/api"
)

// printTable prints schedules in table format
func printTable(schedules []api.Schedule) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "ID\tSCHEDULED AT\tSERVICE\tENVIRONMENTS\tOWNERS\tSTATUS\tDESCRIPTION")
	fmt.Fprintln(w, "----\t------------\t-------\t------------\t------\t------\t-----------")

	for _, s := range schedules {
		desc := ""
		if s.Description != nil {
			desc = truncate(*s.Description, 30)
		}

		// Join environments as comma-separated
		envs := ""
		if len(s.Environments) > 0 {
			for i, env := range s.Environments {
				if i > 0 {
					envs += ","
				}
				envs += string(env)
			}
		}

		// Join owners as comma-separated
		owners := ""
		if len(s.Owners) > 0 {
			for i, owner := range s.Owners {
				if i > 0 {
					owners += ","
				}
				owners += owner
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			truncate(s.Id.String(), 8),
			s.ScheduledAt.Format(time.RFC3339),
			truncate(s.ServiceName, 20),
			truncate(envs, 20),
			truncate(owners, 20),
			s.Status,
			desc,
		)
	}
}

// printJSON prints data in JSON format
func printJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// truncate truncates a string to a maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
