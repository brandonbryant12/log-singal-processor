package logprocessor

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

// PrettyPrintAnomalyInput formats and prints the AnomalyInput with colors for better readability.
func PrettyPrintAnomalyInput(input AnomalyInput) {
	// Define colors
	operationColor := color.New(color.FgCyan, color.Bold)
	tableColor := color.New(color.FgGreen, color.Bold)
	timestampColor := color.New(color.FgYellow)
	headerColor := color.New(color.FgWhite, color.Bold)
	columnColor := color.New(color.FgMagenta)
	signalColor := color.New(color.FgBlue)

	// Print header
	fmt.Println("╔════════════════════════════════════════════════════╗")
	fmt.Printf("║ %-50s ║\n", headerColor.Sprint("ANOMALY DETECTION INPUT"))
	fmt.Println("╠════════════════════════════════════════════════════╣")

	// Print operation and table
	fmt.Printf("║ %-15s %-34s ║\n", headerColor.Sprint("Operation:"), operationColor.Sprint(input.Operation))
	fmt.Printf("║ %-15s %-34s ║\n", headerColor.Sprint("Table:"), tableColor.Sprint(input.Table))

	// Print timestamp
	fmt.Printf("║ %-15s %-34s ║\n", headerColor.Sprint("Timestamp:"), timestampColor.Sprint(input.Timestamp.Format(time.RFC3339)))

	// Print column
	fmt.Printf("║ %-15s %-34s ║\n", headerColor.Sprint("Column:"), columnColor.Sprint(input.Column))

	// Print signal vector
	fmt.Println("╠════════════════════════════════════════════════════╣")
	fmt.Printf("║ %-50s ║\n", headerColor.Sprint("SIGNAL VECTOR"))
	fmt.Println("╠════════════════════════════════════════════════════╣")

	for i, signal := range input.SignalVector {
		// Get the generator name if available
		generatorName := "Unknown"
		if i < len(signalGeneratorNames) {
			generatorName = signalGeneratorNames[i]
		}

		fmt.Printf("║ %-22s %-27s ║\n",
			headerColor.Sprintf("%s:", generatorName),
			signalColor.Sprintf("%.4f", signal))
	}

	// Bottom border
	fmt.Println("╚════════════════════════════════════════════════════╝")
	fmt.Println()
}

// Track generator names for better output labeling
var signalGeneratorNames []string

// RegisterSignalGenerator adds a generator name to the registry for pretty printing
func RegisterSignalGenerator(name string) {
	signalGeneratorNames = append(signalGeneratorNames, name)
}
