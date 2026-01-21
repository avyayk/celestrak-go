package celestrak

import (
	"net/url"
	"strings"
)

// Query represents a Celestrak query that can be used with any endpoint.
// Exactly one of CATNR, INTDES, GROUP, NAME, or SPECIAL must be set.
type Query struct {
	// Query selector (exactly one must be set)
	CATNR   string // Catalog Number (1-9 digits)
	INTDES  string // International Designator (yyyy-nnn)
	GROUP   string // Group name (e.g., "STATIONS", "GPS-OPS")
	NAME    string // Satellite name (partial match)
	SPECIAL string // Special dataset: "GPZ", "GPZ-PLUS", "DECAYING"

	// Format specification (defaults to TLE if not set)
	FORMAT Format

	// Table query flags (only used with table.php endpoint)
	TableFlags TableFlags
}

// TableFlags are optional flags for table.php queries.
type TableFlags struct {
	BSTAR   bool // Show BSTAR value instead of eccentricity
	ShowOps bool // Show operational status flag
	Oldest  bool // Show only objects with data older than 3.5 days
	Docked  bool // Show only docked objects
	Movers  bool // Show only objects drifting >0.1° per day (Active Geosynchronous)
}

// QueryByCATNR creates a query by Catalog Number (NORAD ID).
func QueryByCATNR(catnr string, format Format) Query {
	return Query{CATNR: catnr, FORMAT: format}
}

// QueryByINTDES creates a query by International Designator (yyyy-nnn).
func QueryByINTDES(intdes string, format Format) Query {
	return Query{INTDES: intdes, FORMAT: format}
}

// QueryByGROUP creates a query by group name (e.g., "STATIONS", "GPS-OPS").
func QueryByGROUP(group string, format Format) Query {
	return Query{GROUP: group, FORMAT: format}
}

// QueryByName creates a query by satellite name (partial match).
func QueryByName(name string, format Format) Query {
	return Query{NAME: name, FORMAT: format}
}

// QueryBySPECIAL creates a query for special datasets ("GPZ", "GPZ-PLUS", "DECAYING").
func QueryBySPECIAL(special string, format Format) Query {
	return Query{SPECIAL: special, FORMAT: format}
}

// BuildURL builds a fully-qualified URL for the specified endpoint.
func (q Query) BuildURL(base *url.URL, endpoint string) (string, error) {
	if base == nil {
		return "", &QueryError{Message: "base URL is nil"}
	}
	if endpoint == "" {
		return "", &QueryError{Message: "endpoint is required"}
	}

	key, val, err := q.singleSelector()
	if err != nil {
		return "", err
	}

	format := q.FORMAT
	if format == "" {
		format = FormatTLE
	}

	rel := &url.URL{Path: "/NORAD/elements/" + endpoint}
	params := url.Values{}
	params.Set(key, val)
	params.Set("FORMAT", string(format))

	// Add table flags if this is a table query
	if endpoint == "table.php" {
		q.addTableFlags(&params)
	}

	rel.RawQuery = params.Encode()
	return base.ResolveReference(rel).String(), nil
}

func (q Query) singleSelector() (key, val string, err error) {
	set := func(k, v string) {
		v = strings.TrimSpace(v)
		if v == "" || err != nil {
			return
		}
		if key != "" {
			err = &QueryError{Message: "set exactly one of CATNR/INTDES/GROUP/NAME/SPECIAL"}
			return
		}
		key, val = k, v
	}

	set("CATNR", q.CATNR)
	set("INTDES", q.INTDES)
	set("GROUP", q.GROUP)
	set("NAME", q.NAME)
	set("SPECIAL", q.SPECIAL)

	if err != nil {
		return "", "", err
	}
	if key == "" {
		return "", "", &QueryError{Message: "missing selector (set one of CATNR/INTDES/GROUP/NAME/SPECIAL)"}
	}
	return key, val, nil
}

func (q Query) addTableFlags(params *url.Values) {
	if q.TableFlags.BSTAR {
		params.Set("BSTAR", "1")
	}
	if q.TableFlags.ShowOps {
		params.Set("SHOW-OPS", "1")
	}
	if q.TableFlags.Oldest {
		params.Set("OLDEST", "1")
	}
	if q.TableFlags.Docked {
		params.Set("DOCKED", "1")
	}
	if q.TableFlags.Movers {
		params.Set("MOVERS", "1")
	}
}
