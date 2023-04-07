package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"monitorctl/ddcci"
	"sync"
)

var brightnessCmd = &cobra.Command{
	Use:   "brightness [value]",
	Short: "Set monitor brightness value",
	Long: `Set the monitor brightness value between 0 and 100.
For example:

monitorctl brightness 50`,
	Run: func(cmd *cobra.Command, args []string) {
		brightness := byte(0)

		getValue, _ := cmd.Flags().GetBool("get")
		allMonitors, _ := cmd.Flags().GetBool("all")

		var wg sync.WaitGroup
		if allMonitors {
			i2cBuses, err := ddcci.GetI2CBuses()
			if err != nil {
				log.Fatal(err)
			}
			wg.Add(len(i2cBuses))
			for _, i2cBusNumber := range i2cBuses {
				if getValue {
					brightness, _ := ddcci.GetMonitorBrightness(i2cBusNumber)
					log.Println(brightness)
				} else {
					fmt.Sscanf(args[0], "%d", &brightness)
					go func(bus int) {
						defer wg.Done()
						setErr := ddcci.SetMonitorBrightness(bus, brightness)
						if setErr != nil {
							log.Printf("Failed to set brightness on I2C bus %d: %s", bus, setErr)
						} else {
							fmt.Printf("Brightness set to %d on I2C bus %d\n", brightness, bus)
						}
					}(i2cBusNumber)
				}
			}
		} else {
			i2cBusNumber, _ := cmd.Flags().GetInt("bus")

			wg.Add(1)
			fmt.Sscanf(args[0], "%d", &brightness)
			go func(bus int) {
				defer wg.Done()
				setErr := ddcci.SetMonitorBrightness(bus, brightness)
				if setErr != nil {
					log.Fatal(setErr)
				}
				fmt.Printf("Brightness set to %d on I2C bus %d\n", brightness, bus)
			}(i2cBusNumber)
		}

		wg.Wait()
		fmt.Printf("Monitor brightness set to %d\n", brightness)
	},
}

func init() {
	rootCmd.AddCommand(brightnessCmd)

	brightnessCmd.Flags().BoolP("all", "a", false, "Use all monitors")
	brightnessCmd.Flags().BoolP("get", "g", false, "Get current value")
	brightnessCmd.Flags().IntP("bus", "b", 8, "I2C bus number (default: 8)")
}
