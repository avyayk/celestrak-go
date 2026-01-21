package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/avyayk/celestrak-go/celestrak"
)

func main() {
	ctx := context.Background()

	// Create a client
	client, err := celestrak.NewClient(nil)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Example 1: Fetch ISS (NORAD ID 25544) in TLE format
	fmt.Println("=== Example 1: ISS in TLE format ===")
	issQuery := celestrak.Query{
		CATNR:  "25544",
		FORMAT: celestrak.FormatTLE,
	}
	data, err := client.FetchGP(ctx, issQuery)
	if err != nil {
		handleError("fetching ISS", err)
		return
	}
	fmt.Printf("Fetched %d bytes\n", len(data))
	fmt.Println(string(data[:min(200, len(data))])) // Print first 200 chars
	fmt.Println()

	// Example 2: Fetch ISS in JSON format using helper function
	fmt.Println("=== Example 2: ISS in JSON format ===")
	issJSONQuery := celestrak.QueryByCATNR("25544", celestrak.FormatJSON)
	data, err = client.FetchGP(ctx, issJSONQuery)
	if err != nil {
		handleError("fetching ISS JSON", err)
		return
	}
	fmt.Printf("Fetched %d bytes\n", len(data))
	fmt.Println(string(data[:min(500, len(data))])) // Print first 500 chars
	fmt.Println()

	// Example 3: Fetch Starlink group in JSON-PRETTY format
	fmt.Println("=== Example 3: Starlink group in JSON-PRETTY ===")
	starlinkQuery := celestrak.QueryByGROUP("STARLINK", celestrak.FormatJSONPretty)
	data, err = client.FetchGP(ctx, starlinkQuery)
	if err != nil {
		handleError("fetching Starlink", err)
		return
	}
	fmt.Printf("Fetched %d bytes\n", len(data))
	if len(data) > 0 {
		fmt.Println(string(data[:min(300, len(data))])) // Print first 300 chars
	}
	fmt.Println()

	// Example 4: Fetch by International Designator
	fmt.Println("=== Example 4: Objects from launch 2020-025 ===")
	intdesQuery := celestrak.QueryByINTDES("2020-025", celestrak.FormatJSONPretty)
	data, err = client.FetchGP(ctx, intdesQuery)
	if err != nil {
		handleError("fetching by INTDES", err)
		return
	}
	fmt.Printf("Fetched %d bytes\n", len(data))
	if len(data) > 0 {
		fmt.Println(string(data[:min(300, len(data))]))
	}
	fmt.Println()

	// Example 5: Fetch GEO Protected Zone
	fmt.Println("=== Example 5: GEO Protected Zone ===")
	gpzQuery := celestrak.QueryBySPECIAL("GPZ", celestrak.FormatCSV)
	data, err = client.FetchGP(ctx, gpzQuery)
	if err != nil {
		handleError("fetching GPZ", err)
		return
	}
	fmt.Printf("Fetched %d bytes\n", len(data))
	if len(data) > 0 {
		fmt.Println(string(data[:min(300, len(data))]))
	}
	fmt.Println()

	// Example 6: Fetch first GP data for a launch
	fmt.Println("=== Example 6: First GP data for launch 2024-149 ===")
	firstQuery := celestrak.QueryByINTDES("2024-149", celestrak.FormatJSONPretty)
	data, err = client.FetchGPFirst(ctx, firstQuery)
	if err != nil {
		handleError("fetching first GP", err)
		return
	}
	fmt.Printf("Fetched %d bytes\n", len(data))
	if len(data) > 0 {
		fmt.Println(string(data[:min(300, len(data))]))
	}
	fmt.Println()

	// Example 7: Fetch table data with flags
	fmt.Println("=== Example 7: Stations table with SHOW-OPS flag ===")
	tableQuery := celestrak.Query{
		GROUP: "STATIONS",
		FORMAT: celestrak.FormatXML,
		TableFlags: celestrak.TableFlags{
			ShowOps: true,
		},
	}
	data, err = client.FetchTable(ctx, tableQuery)
	if err != nil {
		handleError("fetching table", err)
		return
	}
	fmt.Printf("Fetched %d bytes\n", len(data))
	if len(data) > 0 {
		fmt.Println(string(data[:min(300, len(data))]))
	}
	fmt.Println()

	// Example 8: Direct query construction (most flexible)
	fmt.Println("=== Example 8: Direct query construction ===")
	directQuery := celestrak.Query{
		NAME:   "COSMOS 2251 DEB",
		FORMAT: celestrak.FormatJSON,
	}
	data, err = client.FetchGP(ctx, directQuery)
	if err != nil {
		handleError("fetching by name", err)
		return
	}
	fmt.Printf("Fetched %d bytes\n", len(data))
	if len(data) > 0 {
		fmt.Println(string(data[:min(300, len(data))]))
	}

	// Print summary
	fmt.Println("\n=== Summary ===")
	fmt.Println("All examples completed. Check output above for results.")
}

// handleError demonstrates proper error handling with type checking
func handleError(operation string, err error) {
	if celestrak.IsErrorResponse(err) {
		if errResp, ok := err.(*celestrak.ErrorResponse); ok {
			if errResp.IsNotFound() {
				log.Printf("Error %s: not found (404)", operation)
			} else if errResp.IsServerError() {
				log.Printf("Error %s: server error (5xx): %v", operation, err)
			} else if errResp.IsRateLimit() {
				log.Printf("Error %s: rate limited (429): %v", operation, err)
			} else {
				log.Printf("Error %s: HTTP error: %v", operation, err)
			}
		}
	} else if celestrak.IsQueryError(err) {
		log.Printf("Error %s: invalid query: %v", operation, err)
	} else {
		log.Printf("Error %s: %v", operation, err)
	}
	os.Exit(1)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

