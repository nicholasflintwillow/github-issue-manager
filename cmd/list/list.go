package list

import (
	"fmt"
	mdparser "github-issue-manager/pkg/mdparser"

	"github.com/spf13/cobra"
)

var folder string

var Cmd = &cobra.Command{
	Use:   "list",
	Short: "List GitHub issues",
	Run: func(cmd *cobra.Command, args []string) {
		files, err := mdparser.ListMarkdownFiles(folder)
		if err != nil {
			fmt.Printf("Error reading folder '%s': %v\n", folder, err)
			return
		}
		for _, file := range files {
			fmt.Println(file)
			frontMatter, err := mdparser.ParseFrontMatter(file)
			if err != nil {
				fmt.Printf("Error parsing front matter in '%s': %v\n", file, err)
				continue
			}
			for key, value := range frontMatter {
				if key == "body" {
					continue
				}
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
	},
}

func init() {
	Cmd.Flags().StringVarP(&folder, "folder", "f", "issues", "Folder containing issue files")
}
