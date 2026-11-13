package upgrade

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Edward-Jackie/lambo/backend/tools/lambo/config"
	"github.com/spf13/cobra"
)

var CmdUpgrade = &cobra.Command{
	Use:     "upgrade",
	Short:   "更新 Lambo 命令行工具",
	Long:    "更新 Lambo 命令行工具",
	Example: "lambo upgrade",
	Args:    cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Printf("go install %s\n", config.LamboCmd)
		cmd := exec.Command("go", "install", config.LamboCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("go install %s: %w", config.LamboCmd, err)
		}
		fmt.Printf("\n🎉 Lambo 升级完成！\n\n")
		return nil
	},
}
