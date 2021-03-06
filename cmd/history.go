// +build ece408ProjectMode

package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/rai-project/auth/provider"
	"github.com/rai-project/client"
	"github.com/rai-project/config"
	"github.com/rai-project/database/mongodb"
	"github.com/spf13/cobra"
	upper "upper.io/db.v3"
	//"gopkg.in/yaml.v2"
)

// rankingCmd represents the ranking command
var historyCmd = &cobra.Command{}

func init() {
	if !ece408ProjectMode {
		return
	}
	historyCmd = &cobra.Command{
		Use:   "history",
		Short: "View history of runs.",
		Long:  `View history of runs associated with user`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Read the profile (e.g. ~/rai_profile.yml)
			prof, err := provider.New()
			if err != nil {
				return err
			}
			// Verify the profile
			ok, err := prof.Verify()
			if err != nil {
				return err
			}
			if !ok {
				return errors.Errorf("cannot authenticate using the credentials in %v", prof.Options().ProfilePath)
			}

			// Create a database  using mongodb with the `config.App.Name` name
			db, err := mongodb.NewDatabase(config.App.Name)
			if err != nil {
				return err
			}
			defer db.Close()

			// Create the Fall2017 collection (mongodb's nomenclature for tables)
			col, err := client.NewEce408JobResponseBodyCollection(db)
			if err != nil {
				return err
			}

			var jobs client.Ece408JobResponseBodys

			condInferencesExist := upper.Cond{"inferences.0 $exists": "true"}
			cond := upper.And(
				condInferencesExist,
				upper.Cond{
					//"is_submission": true,
					"username": prof.Info().Username,
				},
			)

			// find all jobs which are both submissions and have the
			// team name equal to teamname. This would fill the
			// jobs list with the entries found within the collection
			err = col.Find(cond, 0, 0, &jobs)
			if err != nil {
				return err
			}

			if len(jobs) == 0 {
				print("No jobs associated with userid.")
				return nil
			}

			fmt.Println()
			fmt.Println("Last 10 successful runs for user: " + prof.Info().Username)
			fmt.Println()

			// not sure what the heck this is doing
			// TODO: can use a slice
			x := 0
			for _, i := range jobs {
				//Skip items before last 10
				if x > len(jobs)-11 {
					subtag := i.SubmissionTag
					if subtag == "" {
						subtag = "  "
					}
					fmt.Println(subtag + " - " + i.CreatedAt.String() + "\n     " + i.ProjectURL + "\n")
				}
				x++
			}

			return nil
		},
	}
	RootCmd.AddCommand(historyCmd)
}
