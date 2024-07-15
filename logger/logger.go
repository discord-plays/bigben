package logger

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"os"
)

var Logger = log.NewWithOptions(os.Stderr, log.Options{
	ReportCaller:    true,
	ReportTimestamp: true,
	Prefix:          "BigBen",
})

const DevInsertLevel log.Level = 3_065_012

func init() {
	// add a custom style to identify "DEV INSERT" logs
	styles := log.DefaultStyles()
	styles.Levels[DevInsertLevel] = lipgloss.NewStyle().SetString("DEV INSERT").Bold(true).Foreground(lipgloss.Color("201"))
	Logger.SetStyles(styles)
}
