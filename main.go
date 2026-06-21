package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/dharmab/dragonbane-charsheet/internal/character"
	"github.com/dharmab/dragonbane-charsheet/internal/ui"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "dragonbane-char <character.json>",
		Short: "A terminal character sheet for Dragonbane",
		Long:  "A terminal character sheet for Dragonbane.\n\nThe character file is loaded if it exists, or created from a blank sheet if it does not. Changes are saved automatically.",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return []string{"json"}, cobra.ShellCompDirectiveFilterFileExt
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			char, err := character.Load(path)
			if err != nil {
				return fmt.Errorf("loading %s: %w", path, err)
			}

			p := tea.NewProgram(ui.New(char, path))
			if _, err := p.Run(); err != nil {
				return err
			}
			return nil
		},
		SilenceUsage: true,
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
