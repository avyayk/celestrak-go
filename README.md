# celestrak-go

A Go client library for fetching satellite orbital data from [CelesTrak](https://celestrak.org/).

## What is CelesTrak?

CelesTrak provides up-to-date orbital element sets (TLE data) for thousands of satellites. This library makes it easy to fetch this data in various formats (TLE, JSON, XML, CSV) and query by satellite name, NORAD ID, launch, or group.

## Quick Start

### Installation

```bash
go get github.com/avyayk/celestrak-go
```

### Basic Example

Fetching data for the International Space Station:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/avyayk/celestrak-go/celestrak"
)

func main() {
    // Create a client
    client, err := celestrak.NewClient(nil)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    
    // Fetch ISS data (NORAD ID: 25544) in TLE format
    ctx := context.Background()
    query := celestrak.Query{
        CATNR:  "25544",
        FORMAT: celestrak.FormatTLE,
    }
    
    data, err := client.FetchGP(ctx, query)
    if err != nil {
        log.Printf("Error fetching ISS data: %v", err)
        os.Exit(1)
    }
    
    fmt.Println(string(data))
    // Output:
    // ISS (ZARYA)             
    // 1 25544U 98067A   26020.17509289  .00021194  00000+0  38548-3 0  9998
    // 2 25544  51.6334 312.1983 0007785  38.3265 321.8276 15.49442598548811
}
```

## Common Use Cases

### 1. Fetch by NORAD Catalog Number

```go
// Get ISS data in JSON format
query := celestrak.QueryByCATNR("25544", celestrak.FormatJSON)
data, err := client.FetchGP(ctx, query)
```

### 2. Fetch by Satellite Group

```go
// Get all Starlink satellites
query := celestrak.QueryByGROUP("STARLINK", celestrak.FormatJSONPretty)
data, err := client.FetchGP(ctx, query)

// Other groups: "STATIONS", "GPS-OPS", "GALILEO", "IRIDIUM"
```

### 3. Fetch by Launch (International Designator)

```go
// Get all objects from a specific launch
query := celestrak.QueryByINTDES("2020-025", celestrak.FormatJSONPretty)
data, err := client.FetchGP(ctx, query)
```

### 4. Search by Satellite Name

```go
// Find satellites with "COSMOS 2251 DEB" in the name
query := celestrak.QueryByName("COSMOS 2251 DEB", celestrak.FormatJSON)
data, err := client.FetchGP(ctx, query)
```

### 5. Special Datasets

```go
// GEO Protected Zone
query := celestrak.QueryBySPECIAL("GPZ", celestrak.FormatCSV)

// GEO Protected Zone Plus
query := celestrak.QueryBySPECIAL("GPZ-PLUS", celestrak.FormatJSONPretty)

// Potential Decays
query := celestrak.QueryBySPECIAL("DECAYING", celestrak.FormatJSONPretty)
```

## Available Formats

The library supports all CelesTrak formats:

- `FormatTLE` - Three-line element sets (default)
- `Format3LE` - Three-line with satellite name
- `Format2LE` - Two-line element sets
- `FormatJSON` - JSON format
- `FormatJSONPretty` - Pretty-printed JSON
- `FormatXML` - CCSDS OMM XML format
- `FormatKVN` - CCSDS OMM KVN format
- `FormatCSV` - CSV format

## Different Endpoints

### Current GP Data (`FetchGP`)

```go
// Get current orbital data
data, err := client.FetchGP(ctx, query)
```

### First GP Data (`FetchGPFirst`)

```go
// Get the first GP data available for a launch
query := celestrak.QueryByINTDES("2024-149", celestrak.FormatJSONPretty)
data, err := client.FetchGPFirst(ctx, query)
```

### Last GP Data (`FetchGPLast`)

```go
// Get the most recent GP data
data, err := client.FetchGPLast(ctx, query)
```

### Table Data (`FetchTable`)

```go
// Get table data with optional flags
query := celestrak.Query{
    GROUP:  "STATIONS",
    FORMAT: celestrak.FormatXML,
    TableFlags: celestrak.TableFlags{
        ShowOps: true,  // Show operational status
        Oldest:  true,  // Show only objects with data older than 3.5 days
    },
}
data, err := client.FetchTable(ctx, query)
```

**Table Flags:**
- `BSTAR` - Show BSTAR value instead of eccentricity
- `ShowOps` - Show operational status flag
- `Oldest` - Show only objects with data older than 3.5 days
- `Docked` - Show only docked objects
- `Movers` - Show only objects drifting >0.1° per day

## Error Handling

The library returns errors that should be handled by the caller. Use Go's standard `log` package or structured logging with `log/slog` (Go 1.21+).

### Basic Error Handling

```go
import (
    "log"
    "os"
)

data, err := client.FetchGP(ctx, query)
if err != nil {
    if celestrak.IsErrorResponse(err) {
        errResp := err.(*celestrak.ErrorResponse)
        
        if errResp.IsNotFound() {
            log.Printf("Satellite not found: %v", err)
            return
        } else if errResp.IsServerError() {
            log.Printf("Server error (will retry): %v", err)
            return
        } else if errResp.IsRateLimit() {
            log.Printf("Rate limited: %v", err)
            return
        }
    } else if celestrak.IsQueryError(err) {
        log.Printf("Invalid query: %v", err)
        return
    }
    
    log.Printf("Error fetching data: %v", err)
    return
}
```

### Structured Logging with log/slog

For production applications, use structured logging:

```go
import (
    "log/slog"
    "os"
)

logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

data, err := client.FetchGP(ctx, query)
if err != nil {
    if celestrak.IsErrorResponse(err) {
        errResp := err.(*celestrak.ErrorResponse)
        
        if errResp.IsNotFound() {
            logger.Warn("Satellite not found",
                "error", err,
                "status", errResp.Response.StatusCode,
            )
            return
        } else if errResp.IsServerError() {
            logger.Error("Server error",
                "error", err,
                "status", errResp.Response.StatusCode,
            )
            return
        } else if errResp.IsRateLimit() {
            logger.Warn("Rate limited",
                "error", err,
                "status", errResp.Response.StatusCode,
            )
            return
        }
    } else if celestrak.IsQueryError(err) {
        logger.Error("Invalid query", "error", err)
        return
    }
    
    logger.Error("Failed to fetch data", "error", err)
    return
}
```

### Error Handling Best Practices

- **Don't ignore errors**: Always check and handle `err` return values
- **Use appropriate log levels**: `Error` for failures, `Warn` for recoverable issues, `Info` for normal operations
- **Include context**: Log relevant information (query parameters, status codes, etc.)
- **Return errors up the stack**: In libraries or when you can't handle the error, return it rather than logging
- **Use `log.Fatal` sparingly**: Only use `log.Fatal` in `main()` when the program cannot continue

## Advanced Features

### Automatic Retries

The library automatically retries transient failures (network errors, server errors) with exponential backoff:

```go
client, _ := celestrak.NewClient(nil)

// Configure retry behavior (optional)
// Default: 3 retries, 1 second initial delay
client = client.WithRetries(5, 2*time.Second)
```

### Context Support

Use contexts for timeouts and cancellation:

```go
// Set a 30-second timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

data, err := client.FetchGP(ctx, query)
```

### Caching (ETag Support)

Implement the `Cache` interface for automatic ETag-based caching:

```go
type Cache interface {
    Get(key string) (data []byte, etag string, ok bool)
    Put(key string, data []byte, etag string)
}

// Use with client
client = client.WithCache(myCache)
```

### Custom HTTP Client

```go
customClient := &http.Client{
    Timeout: 60 * time.Second,
    // Add your own transport, etc.
}

celestrakClient, _ := celestrak.NewClient(customClient)
```

### Custom User-Agent

```go
client = client.WithUserAgent("my-app/1.0")
```

## Complete Example

Example demonstrating multiple features:

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/avyayk/celestrak-go/celestrak"
)

func main() {
    // Create client with timeout
    client, err := celestrak.NewClient(nil)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    
    // Configure retries
    client = client.WithRetries(3, 1*time.Second)
    
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Fetch ISS data in JSON format
    query := celestrak.QueryByCATNR("25544", celestrak.FormatJSON)
    data, err := client.FetchGP(ctx, query)
    if err != nil {
        log.Printf("Error fetching ISS data: %v", err)
        os.Exit(1)
    }
    
    // Parse JSON response
    var satellites []map[string]interface{}
    if err := json.Unmarshal(data, &satellites); err != nil {
        log.Printf("Error parsing JSON: %v", err)
        os.Exit(1)
    }
    
    // Print satellite info
    for _, sat := range satellites {
        fmt.Printf("Name: %s\n", sat["OBJECT_NAME"])
        fmt.Printf("NORAD ID: %.0f\n", sat["NORAD_CAT_ID"])
        fmt.Printf("Inclination: %.4f°\n", sat["INCLINATION"])
        fmt.Println()
    }
}
```

## Best Practices

1. **Respect Rate Limits**: CelesTrak recommends limiting requests to 3-4 times per day. Use ETag caching to reduce requests.

2. **Use Contexts**: Use contexts for cancellation and timeouts in production code.

3. **Handle Errors**: Check error types to handle different failure modes appropriately.

4. **Choose the Right Format**: 
   - `JSON` or `JSONPretty` for programmatic access
   - `TLE` for compatibility with existing TLE parsers
   - `CSV` for spreadsheet analysis

5. **Cache When Possible**: Implement caching to reduce load on CelesTrak and improve performance.

## Troubleshooting

**"context must be non-nil"**
- Pass a valid context: `context.Background()` or `context.WithTimeout()`

**"Query: missing selector"**
- Set exactly one of: `CATNR`, `INTDES`, `GROUP`, `NAME`, or `SPECIAL`

**"Query: set exactly one of..."**
- Multiple selectors were set. Only one is allowed.

**Empty responses**
- Some queries may return empty results if no satellites match. Check the response length.

## Resources

- [CelesTrak Documentation](https://celestrak.org/NORAD/documentation/)
- [CelesTrak Current Data](https://celestrak.org/NORAD/elements/)
- [Example Code](./example/celestrak.go)

## Contributing

Contributions welcome. Open issues or submit pull requests.

## License

[Add your license here]

