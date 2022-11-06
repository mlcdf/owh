package commands

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/config"
)

func Tasks(client *api.Client, hosting string) error {
	if hosting == "" {
		link, err := config.EnsureLink()
		if err != nil {
			if err != config.ErrFolderNotLinked {
				return err
			}
			fmt.Println("Folder not link. Please run: owh link first")
		}
		hosting = link.Hosting
	}

	tasks, err := client.Tasks(hosting)
	if err != nil {
		return err
	}

	idColumn := lipgloss.NewStyle().Width(12)
	statusColumn := lipgloss.NewStyle().Width(12)
	functionColumn := lipgloss.NewStyle().Width(30)
	startDateColumn := lipgloss.NewStyle().Width(36)

	fmt.Printf(
		"%s %s %s %s %s\n",
		idColumn.Render("ID"),
		statusColumn.Render("STATUS"),
		functionColumn.Render("FUNCTION"),
		startDateColumn.Render("START DATE"),
		"LAST UPDATE",
	)

	for _, task := range tasks {
		fmt.Println(idColumn.Render(fmt.Sprintf("%d", task.ID)), statusColumn.Render(task.Status), functionColumn.Render(task.Function), startDateColumn.Render(task.StartDate.String()), task.LastUpdate)
	}

	return nil
}
