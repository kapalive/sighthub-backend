// Quick SPIn API test — run when P1 terminal is online:
//   go run cmd/testfind/pos_spin_test.go
package main

import (
	"fmt"
	"os"
	"sighthub-backend/pkg/spin"
)

func main() {
	c := spin.NewClient(
		spin.SandboxBaseURL, // https://test.spinpos.net/spin
		"HUjca5uQWx",       // AuthKey
		"118126783745",      // TPN (physical P1)
		0,                   // default timeout 130s
	)

	// 1. Check terminal status
	fmt.Println("=== Terminal Status ===")
	resp, err := c.TerminalStatus()
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
	fmt.Printf("ResultCode: %d, Message: %s\n", resp.ResultCode, resp.Message)
	fmt.Printf("Raw: %s\n\n", resp.RawJSON)

	// 2. If online — try a $1.00 sale
	if resp.ResultCode != 0 {
		fmt.Println("Terminal is OFFLINE. Power on the P1 and connect to WiFi.")
		fmt.Println("Then re-run this test.")
		os.Exit(0)
	}

	fmt.Println("=== Sale $1.00 ===")
	saleResp, err := c.Sale(1.00, "SIGHTHUB_TEST_001", "INV-TEST")
	if err != nil {
		fmt.Println("SALE ERROR:", err)
		os.Exit(1)
	}
	fmt.Printf("ResultCode: %d, StatusCode: %s\n", saleResp.ResultCode, saleResp.StatusCode)
	fmt.Printf("Message: %s\n", saleResp.Message)
	fmt.Printf("AuthCode: %s\n", saleResp.AuthCode)
	fmt.Printf("Approved: %v\n", saleResp.IsApproved())
	if saleResp.CardData != nil {
		fmt.Printf("Card: %s ****%s (%s)\n", saleResp.CardData.CardType, saleResp.CardData.Last4, saleResp.CardData.EntryMethod)
	}
	fmt.Printf("\nRaw: %s\n", saleResp.RawJSON)
}
