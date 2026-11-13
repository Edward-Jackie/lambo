package command

import (
	"github.com/Edward-Jackie/lambo/config"
	"github.com/Edward-Jackie/lambo/internal/command/create"
	newproject "github.com/Edward-Jackie/lambo/internal/command/new"
	"github.com/Edward-Jackie/lambo/internal/command/run"
	"github.com/Edward-Jackie/lambo/internal/command/upgrade"
	"github.com/Edward-Jackie/lambo/internal/command/wire"
	"github.com/spf13/cobra"
)

var CmdRoot = &cobra.Command{
	Use: "lambo", Short: "Lambo：Go + Vben 业务脚手架开发工具",
	Example: "lambo create frontend category --schema examples/category.json --dry-run",
	Version: config.Version, SilenceErrors: true, SilenceUsage: true,
}

func init() {
	CmdRoot.AddCommand(newproject.CmdNew, create.CmdCreate, run.CmdRun, wire.CmdWire, upgrade.CmdUpgrade)
	wire.CmdWire.AddCommand(wire.CmdWireAll)
	CmdRoot.CompletionOptions.DisableDefaultCmd = true
}

func Execute() error { return CmdRoot.Execute() }
