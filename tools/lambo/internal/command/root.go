package command

import (
	"github.com/Edward-Jackie/lambo/tools/lambo/config"
	"github.com/Edward-Jackie/lambo/tools/lambo/internal/command/create"
	"github.com/Edward-Jackie/lambo/tools/lambo/internal/command/db"
	newproject "github.com/Edward-Jackie/lambo/tools/lambo/internal/command/new"
	"github.com/Edward-Jackie/lambo/tools/lambo/internal/command/run"
	"github.com/Edward-Jackie/lambo/tools/lambo/internal/command/upgrade"
	"github.com/Edward-Jackie/lambo/tools/lambo/internal/command/wire"
	"github.com/spf13/cobra"
)

var CmdRoot = &cobra.Command{
	Use: "lambo", Short: "Lambo：Go + Vben 业务脚手架开发工具",
	Example: "lambo create frontend category --schema tools/lambo/examples/category.json --dry-run",
	Version: config.Version, SilenceErrors: true, SilenceUsage: true,
}

func init() {
	CmdRoot.AddCommand(newproject.CmdNew, create.CmdCreate, db.CmdDB, run.CmdRun, wire.CmdWire, upgrade.CmdUpgrade)
	wire.CmdWire.AddCommand(wire.CmdWireAll)
	CmdRoot.CompletionOptions.DisableDefaultCmd = true
}

func Execute() error { return CmdRoot.Execute() }
