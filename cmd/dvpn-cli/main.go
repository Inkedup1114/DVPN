package main

import (
    "fmt"
    "log"
    
    "github.com/spf13/cobra"
)

func main() {
    var rootCmd = &cobra.Command{
        Use:   "dvpn-cli",
        Short: "DVPN command line interface",
    }
    
    // Status command
    var statusCmd = &cobra.Command{
        Use:   "status",
        Short: "Show node status",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("DVPN Node Status:")
            fmt.Println("  Status: Running")
            fmt.Println("  Block Height: 1")
            fmt.Println("  Peers: 0")
            fmt.Println("  Mining: Disabled")
            fmt.Println("  VPN: Disabled")
        },
    }
    
    // Version command
    var versionCmd = &cobra.Command{
        Use:   "version",
        Short: "Show version information",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("DVPN CLI v0.1.0")
        },
    }
    
    rootCmd.AddCommand(statusCmd, versionCmd)
    
    if err := rootCmd.Execute(); err != nil {
        log.Fatal(err)
    }
}
