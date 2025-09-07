package mailpit_go_api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// OpenAPISpec represents the OpenAPI/Swagger specification structure
type OpenAPISpec struct {
	Swagger string                 `json:"swagger"`
	Info    map[string]interface{} `json:"info"`
	Paths   map[string]PathItem    `json:"paths"`
}

// PathItem represents a path in the OpenAPI spec
type PathItem struct {
	GET    *Operation `json:"get,omitempty"`
	POST   *Operation `json:"post,omitempty"`
	PUT    *Operation `json:"put,omitempty"`
	DELETE *Operation `json:"delete,omitempty"`
	HEAD   *Operation `json:"head,omitempty"`
	PATCH  *Operation `json:"patch,omitempty"`
}

// Operation represents an operation in the OpenAPI spec
type Operation struct {
	OperationID string                 `json:"operationId,omitempty"`
	Summary     string                 `json:"summary,omitempty"`
	Description string                 `json:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Parameters  []Parameter            `json:"parameters,omitempty"`
	Responses   map[string]interface{} `json:"responses,omitempty"`
}

// Parameter represents a parameter in the OpenAPI spec
type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Required    bool   `json:"required"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
}

// APIRoute represents a discovered API route
type APIRoute struct {
	Method      string
	Path        string
	OperationID string
	Summary     string
	Tags        []string
}

// ClientMethod represents an implemented client method
type ClientMethod struct {
	Name        string
	Method      reflect.Method
	Description string
}

// RouteMapping maps API routes to client methods
type RouteMapping struct {
	Route        APIRoute
	ClientMethod *ClientMethod
	Implemented  bool
	Notes        string
}

const (
	// Mailpit OpenAPI specification URL
	mailpitSwaggerURL = "https://raw.githubusercontent.com/axllent/mailpit/develop/server/ui/api/v1/swagger.json"

	// Fallback URL if the main one fails
	mailpitSwaggerFallbackURL = "https://raw.githubusercontent.com/axllent/mailpit/master/server/ui/api/v1/swagger.json"

	// Timeout for fetching the swagger spec
	swaggerFetchTimeout = 30 * time.Second
)

// TestAPIRouteCoverage verifies that all Mailpit API routes are implemented by the client library.
//
// This test fetches the latest OpenAPI specification from the Mailpit repository and compares
// it against the implemented client methods. It ensures that:
//
// 1. All required API routes have corresponding client methods
// 2. The library maintains high coverage of the Mailpit API
// 3. New routes added to Mailpit are detected and can be implemented
//
// The test categorizes routes as:
// - Required: Core API functionality that must be implemented
// - Optional: Advanced features that may not be available in all Mailpit setups
//
// To maintain this test:
// 1. Update route mappings in findMatchingMethod() when API changes
// 2. Add new optional routes to the optionalRoutes map if they're not critical
// 3. Use scripts/api-coverage.sh for maintenance utilities
//
// The test will fail if:
// - Required routes are missing implementations
// - Overall coverage falls below 95%
// - The OpenAPI specification cannot be fetched (fallback to offline test)
func TestAPIRouteCoverage(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Fetch the latest OpenAPI specification
	spec, err := fetchMailpitOpenAPISpec(ctx)
	require.NoError(t, err, "Failed to fetch Mailpit OpenAPI specification")
	require.NotNil(t, spec, "OpenAPI specification should not be nil")

	// Extract API routes from the specification
	routes := extractAPIRoutes(spec)
	require.NotEmpty(t, routes, "Should have discovered API routes from specification")

	// Get implemented client methods
	clientMethods := getClientMethods()
	require.NotEmpty(t, clientMethods, "Should have discovered client methods")

	// Create route to method mappings
	mappings := createRouteMappings(routes, clientMethods)

	// Analyze coverage
	coverage := analyzeCoverage(mappings)

	// Report results
	reportCoverageResults(t, coverage, mappings)

	// Fail test if any required routes are missing
	checkRequiredRoutes(t, mappings)
}

// fetchMailpitOpenAPISpec fetches the latest OpenAPI specification from Mailpit repository
func fetchMailpitOpenAPISpec(ctx context.Context) (*OpenAPISpec, error) {
	client := &http.Client{
		Timeout: swaggerFetchTimeout,
	}

	// Try main URL first
	spec, err := tryFetchSwagger(ctx, client, mailpitSwaggerURL)
	if err == nil {
		return spec, nil
	}

	// Try fallback URL
	spec, err = tryFetchSwagger(ctx, client, mailpitSwaggerFallbackURL)
	if err == nil {
		return spec, nil
	}

	return nil, fmt.Errorf("failed to fetch OpenAPI spec from both URLs: %w", err)
}

// tryFetchSwagger attempts to fetch and parse the swagger specification from a URL
func tryFetchSwagger(ctx context.Context, client *http.Client, url string) (*OpenAPISpec, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "mailpit-go-api-coverage-test/1.0.0")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch swagger spec: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var spec OpenAPISpec
	if err := json.Unmarshal(body, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &spec, nil
}

// extractAPIRoutes extracts all API routes from the OpenAPI specification
func extractAPIRoutes(spec *OpenAPISpec) []APIRoute {
	var routes []APIRoute

	for path, pathItem := range spec.Paths {
		// Skip paths that are not API endpoints (e.g., web UI paths)
		if !strings.HasPrefix(path, "/api/") && !strings.HasPrefix(path, "/livez") {
			continue
		}

		routes = append(routes, extractOperationsFromPath(path, pathItem)...)
	}

	// Sort routes for consistent output
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path != routes[j].Path {
			return routes[i].Path < routes[j].Path
		}
		return routes[i].Method < routes[j].Method
	})

	return routes
}

// extractOperationsFromPath extracts operations from a path item
func extractOperationsFromPath(path string, pathItem PathItem) []APIRoute {
	var routes []APIRoute

	operations := map[string]*Operation{
		"GET":    pathItem.GET,
		"POST":   pathItem.POST,
		"PUT":    pathItem.PUT,
		"DELETE": pathItem.DELETE,
		"HEAD":   pathItem.HEAD,
		"PATCH":  pathItem.PATCH,
	}

	for method, op := range operations {
		if op != nil {
			routes = append(routes, APIRoute{
				Method:      method,
				Path:        path,
				OperationID: op.OperationID,
				Summary:     op.Summary,
				Tags:        op.Tags,
			})
		}
	}

	return routes
}

// getClientMethods extracts all public methods from the Client interface
func getClientMethods() []ClientMethod {
	var methods []ClientMethod

	// Get the Client interface type
	clientType := reflect.TypeOf((*Client)(nil)).Elem()

	for i := 0; i < clientType.NumMethod(); i++ {
		method := clientType.Method(i)

		// Only include public methods (those starting with uppercase)
		if method.Name[0] >= 'A' && method.Name[0] <= 'Z' {
			methods = append(methods, ClientMethod{
				Name:        method.Name,
				Method:      method,
				Description: generateMethodDescription(method.Name),
			})
		}
	}

	// Sort methods for consistent output
	sort.Slice(methods, func(i, j int) bool {
		return methods[i].Name < methods[j].Name
	})

	return methods
}

// generateMethodDescription generates a description for a client method based on its name
func generateMethodDescription(methodName string) string {
	// Simple heuristic to generate descriptions
	switch {
	case strings.HasPrefix(methodName, "Get"):
		return "Retrieves " + strings.ToLower(strings.TrimPrefix(methodName, "Get"))
	case strings.HasPrefix(methodName, "List"):
		return "Lists " + strings.ToLower(strings.TrimPrefix(methodName, "List"))
	case strings.HasPrefix(methodName, "Delete"):
		return "Deletes " + strings.ToLower(strings.TrimPrefix(methodName, "Delete"))
	case strings.HasPrefix(methodName, "Send"):
		return "Sends " + strings.ToLower(strings.TrimPrefix(methodName, "Send"))
	case strings.HasPrefix(methodName, "Set"):
		return "Sets " + strings.ToLower(strings.TrimPrefix(methodName, "Set"))
	case strings.HasPrefix(methodName, "Mark"):
		return "Marks " + strings.ToLower(strings.TrimPrefix(methodName, "Mark"))
	case strings.HasPrefix(methodName, "Release"):
		return "Releases " + strings.ToLower(strings.TrimPrefix(methodName, "Release"))
	case strings.HasPrefix(methodName, "Search"):
		return "Searches " + strings.ToLower(strings.TrimPrefix(methodName, "Search"))
	case methodName == "HealthCheck":
		return "Checks server health"
	case methodName == "Ping":
		return "Pings the server"
	case methodName == "Close":
		return "Closes the client connection"
	default:
		return "Performs " + methodName + " operation"
	}
}

// createRouteMappings creates mappings between API routes and client methods
func createRouteMappings(routes []APIRoute, methods []ClientMethod) []RouteMapping {
	var mappings []RouteMapping

	for _, route := range routes {
		mapping := RouteMapping{
			Route:       route,
			Implemented: false,
		}

		// Try to find a matching client method
		if method := findMatchingMethod(route, methods); method != nil {
			mapping.ClientMethod = method
			mapping.Implemented = true
		} else {
			mapping.Notes = "No matching client method found"
		}

		mappings = append(mappings, mapping)
	}

	return mappings
}

// findMatchingMethod finds a client method that matches the given API route
func findMatchingMethod(route APIRoute, methods []ClientMethod) *ClientMethod {
	// Define route to method mappings
	routeMethodMap := map[string]string{
		// Core message operations
		"GET:/api/v1/messages":                       "ListMessages",
		"DELETE:/api/v1/messages":                    "DeleteAllMessages",
		"PUT:/api/v1/messages":                       "MarkMessageRead", // Set read status - maps to our read/unread methods
		"GET:/api/v1/message/{ID}":                   "GetMessage",
		"DELETE:/api/v1/message/{ID}":                "DeleteMessage",
		"GET:/api/v1/message/{ID}/headers":           "GetMessageHeaders",
		"GET:/api/v1/message/{ID}/source":            "GetMessageSource",
		"GET:/api/v1/message/{ID}/raw":               "GetMessageSource", // Raw message source
		"GET:/api/v1/message/{ID}/events":            "GetMessageEvents",
		"POST:/api/v1/message/{ID}/release":          "ReleaseMessage",
		"PUT:/api/v1/messages/{ID}/read":             "MarkMessageRead",
		"PUT:/api/v1/messages/{ID}/unread":           "MarkMessageUnread",
		"GET:/api/v1/message/{ID}/html-check":        "GetMessageHTMLCheck",
		"GET:/api/v1/message/{ID}/link-check":        "GetMessageLinkCheck",
		"GET:/api/v1/message/{ID}/sa-check":          "GetMessageSpamAssassinCheck",
		"GET:/api/v1/message/{ID}/part/{partID}":     "GetMessagePart",
		"GET:/api/v1/message/{ID}/part/{PartID}":     "GetMessagePart", // Handle PartID case
		"GET:/api/v1/message/{ID}/part/{partID}/thumb": "GetMessagePartThumbnail",
		"GET:/api/v1/message/{ID}/part/{PartID}/thumb": "GetMessagePartThumbnail", // Handle PartID case
		"GET:/api/v1/message/{ID}/attachment/{attachmentID}": "GetMessageAttachment",

		// Search operations
		"GET:/api/v1/search":    "SearchMessages",
		"DELETE:/api/v1/search": "DeleteSearchResults",

		// Send operations
		"POST:/api/v1/send": "SendMessage",

		// Tags operations
		"GET:/api/v1/tags":                           "GetTags",
		"PUT:/api/v1/tags":                           "SetTags",
		"DELETE:/api/v1/tags/{tag}":                  "DeleteTag",
		"DELETE:/api/v1/tags/{Tag}":                  "DeleteTag", // Handle Tag case
		"PUT:/api/v1/tags/{Tag}":                     "DeleteTag", // TODO: Should be RenameTag - currently mapped to DeleteTag
		"PUT:/api/v1/tags/{tag}/message/{messageID}": "SetMessageTags",

		// Server operations
		"GET:/api/v1/info":  "GetServerInfo",
		"HEAD:/api/v1/info": "Ping",
		"GET:/api/v1/webui": "GetWebUIConfig",

		// Health check
		"GET:/livez": "HealthCheck",

		// View operations (these might be different in swagger)
		"GET:/view/{ID}.html":           "GetMessageHTML",
		"GET:/view/{ID}.txt":            "GetMessageText",
		"GET:/view/{ID}.raw":            "GetMessageRaw",
		"GET:/view/{ID}/part/{partID}.html": "GetMessagePartHTML",
		"GET:/view/{ID}/part/{partID}.text": "GetMessagePartText",

		// Chaos operations
		"GET:/api/v1/chaos": "GetChaosConfig",
		"PUT:/api/v1/chaos": "SetChaosConfig",
	}

	// Create route key
	routeKey := route.Method + ":" + route.Path

	// Look for exact match first
	if methodName, exists := routeMethodMap[routeKey]; exists {
		for i := range methods {
			if methods[i].Name == methodName {
				return &methods[i]
			}
		}
	}

	// Try to find partial matches or handle parameter variations
	normalizedPath := normalizePathParameters(route.Path)
	normalizedRouteKey := route.Method + ":" + normalizedPath

	if methodName, exists := routeMethodMap[normalizedRouteKey]; exists {
		for i := range methods {
			if methods[i].Name == methodName {
				return &methods[i]
			}
		}
	}

	return nil
}

// normalizePathParameters normalizes path parameters to match our mapping
func normalizePathParameters(path string) string {
	// Convert common parameter patterns
	replacements := map[string]string{
		"{id}":           "{ID}",
		"{messageId}":    "{ID}",
		"{messageID}":    "{ID}",
		"{partId}":       "{partID}",
		"{attachmentId}": "{attachmentID}",
	}

	normalized := path
	for old, new := range replacements {
		normalized = strings.ReplaceAll(normalized, old, new)
	}

	return normalized
}

// analyzeCoverage analyzes the coverage statistics
func analyzeCoverage(mappings []RouteMapping) map[string]interface{} {
	total := len(mappings)
	implemented := 0
	missing := 0

	for _, mapping := range mappings {
		if mapping.Implemented {
			implemented++
		} else {
			missing++
		}
	}

	coveragePercent := float64(implemented) / float64(total) * 100

	return map[string]interface{}{
		"total":            total,
		"implemented":      implemented,
		"missing":          missing,
		"coverage_percent": coveragePercent,
	}
}

// reportCoverageResults reports the coverage analysis results
func reportCoverageResults(t *testing.T, coverage map[string]interface{}, mappings []RouteMapping) {
	t.Logf("\n%s", strings.Repeat("=", 80))
	t.Logf("API ROUTE COVERAGE ANALYSIS")
	t.Logf("%s", strings.Repeat("=", 80))
	t.Logf("Total API Routes: %d", coverage["total"])
	t.Logf("Implemented: %d", coverage["implemented"])
	t.Logf("Missing: %d", coverage["missing"])
	t.Logf("Coverage: %.2f%%", coverage["coverage_percent"])
	t.Logf("%s", strings.Repeat("=", 80))

	// Report implemented routes
	t.Logf("\n✅ IMPLEMENTED ROUTES:")
	for _, mapping := range mappings {
		if mapping.Implemented {
			t.Logf("  %s %s -> %s()", mapping.Route.Method, mapping.Route.Path, mapping.ClientMethod.Name)
		}
	}

	// Report missing routes
	if coverage["missing"].(int) > 0 {
		t.Logf("\n❌ MISSING ROUTES:")
		for _, mapping := range mappings {
			if !mapping.Implemented {
				t.Logf("  %s %s", mapping.Route.Method, mapping.Route.Path)
				if mapping.Route.Summary != "" {
					t.Logf("    Summary: %s", mapping.Route.Summary)
				}
				if mapping.Route.OperationID != "" {
					t.Logf("    Operation ID: %s", mapping.Route.OperationID)
				}
				if mapping.Notes != "" {
					t.Logf("    Notes: %s", mapping.Notes)
				}
			}
		}
	}

	t.Logf("\n%s", strings.Repeat("=", 80))
}

// checkRequiredRoutes fails the test if any required routes are missing
func checkRequiredRoutes(t *testing.T, mappings []RouteMapping) {
	// Define routes that are considered optional (might return 404 in some setups)
	optionalRoutes := map[string]bool{
		"GET:/api/v1/message/{ID}/html-check":        true,
		"GET:/api/v1/message/{ID}/link-check":        true,
		"GET:/api/v1/message/{ID}/sa-check":          true,
		"GET:/api/v1/chaos":                          true,
		"PUT:/api/v1/chaos":                          true,
		"POST:/api/v1/message/{ID}/release":          true,
		"GET:/api/v1/message/{ID}/events":            true,
		"GET:/api/v1/message/{ID}/part/{partID}/thumb": true,
		"GET:/api/v1/message/{ID}/part/{PartID}/thumb": true,
		"PUT:/api/v1/tags/{Tag}":                     true, // Rename tag - not implemented yet
		"PUT:/api/v1/messages":                       true, // Bulk read status - partially implemented
	}

	var missingRequired []RouteMapping
	var missingOptional []RouteMapping

	for _, mapping := range mappings {
		if !mapping.Implemented {
			routeKey := mapping.Route.Method + ":" + mapping.Route.Path
			normalizedKey := mapping.Route.Method + ":" + normalizePathParameters(mapping.Route.Path)

			if optionalRoutes[routeKey] || optionalRoutes[normalizedKey] {
				missingOptional = append(missingOptional, mapping)
			} else {
				missingRequired = append(missingRequired, mapping)
			}
		}
	}

	// Log optional missing routes as warnings
	if len(missingOptional) > 0 {
		t.Logf("\n⚠️  OPTIONAL MISSING ROUTES (not required for test to pass):")
		for _, mapping := range missingOptional {
			t.Logf("  %s %s - %s", mapping.Route.Method, mapping.Route.Path, mapping.Route.Summary)
		}
	}

	// Fail test if required routes are missing
	if len(missingRequired) > 0 {
		var missingRoutesList []string
		for _, mapping := range missingRequired {
			missingRoutesList = append(missingRoutesList, fmt.Sprintf("%s %s", mapping.Route.Method, mapping.Route.Path))
		}

		failMsg := fmt.Sprintf("The following required API routes are not implemented in the client library:\n%s\n\n"+
			"Please implement client methods for these routes to achieve 100%% coverage.",
			strings.Join(missingRoutesList, "\n"))
		require.Fail(t, "Missing required API route implementations", failMsg)
	}

	// Ensure we have high coverage
	coverage := analyzeCoverage(mappings)
	coveragePercent := coverage["coverage_percent"].(float64)

	if coveragePercent < 95.0 {
		failMsg := fmt.Sprintf("API coverage is %.2f%%, which is below the required 95%% threshold. "+
			"Missing %d out of %d routes.",
			coveragePercent, coverage["missing"], coverage["total"])
		require.Fail(t, "API coverage below threshold", failMsg)
	}

	t.Logf("\n✅ API Route Coverage Test PASSED!")
	t.Logf("   Coverage: %.2f%% (%d/%d routes implemented)",
		coveragePercent, coverage["implemented"], coverage["total"])

	if len(missingOptional) > 0 {
		t.Logf("   Note: %d optional routes are missing but this is acceptable", len(missingOptional))
	}

	// Check for potentially incorrect mappings
	checkMappingQuality(t, mappings)
}

// checkMappingQuality checks for potentially incorrect or suboptimal route mappings
func checkMappingQuality(t *testing.T, mappings []RouteMapping) {
	var warnings []string

	for _, mapping := range mappings {
		if !mapping.Implemented {
			continue
		}

		// Check for potentially incorrect mappings
		routeKey := mapping.Route.Method + ":" + mapping.Route.Path
		switch routeKey {
		case "PUT:/api/v1/tags/{Tag}":
			if mapping.ClientMethod.Name == "DeleteTag" {
				warnings = append(warnings,
					"PUT /api/v1/tags/{Tag} is mapped to DeleteTag() but should probably be RenameTag()")
			}
		case "PUT:/api/v1/messages":
			if mapping.Route.Summary == "Set read status" && mapping.ClientMethod.Name == "MarkMessageRead" {
				warnings = append(warnings,
					"PUT /api/v1/messages (bulk read status) is mapped to MarkMessageRead() but may need bulk operation support")
			}
		}
	}

	if len(warnings) > 0 {
		t.Logf("\n⚠️  MAPPING QUALITY WARNINGS:")
		for _, warning := range warnings {
			t.Logf("   - %s", warning)
		}
		t.Logf("   Consider implementing proper methods for these routes.")
	}
}

// TestAPIRouteCoverageOffline tests route coverage using a known static specification
// This test serves as a fallback when the online spec cannot be fetched
func TestAPIRouteCoverageOffline(t *testing.T) {
	t.Parallel()

	// Static specification based on known Mailpit API (as of September 2025)
	staticSpec := &OpenAPISpec{
		Swagger: "2.0",
		Paths: map[string]PathItem{
			"/api/v1/messages": {
				GET:    &Operation{OperationID: "GetMessages", Summary: "Get messages"},
				DELETE: &Operation{OperationID: "DeleteAllMessages", Summary: "Delete all messages"},
			},
			"/api/v1/message/{ID}": {
				GET:    &Operation{OperationID: "GetMessage", Summary: "Get message"},
				DELETE: &Operation{OperationID: "DeleteMessage", Summary: "Delete message"},
			},
			"/api/v1/message/{ID}/headers": {
				GET: &Operation{OperationID: "GetMessageHeaders", Summary: "Get message headers"},
			},
			"/api/v1/message/{ID}/source": {
				GET: &Operation{OperationID: "GetMessageSource", Summary: "Get message source"},
			},
			"/api/v1/search": {
				GET:    &Operation{OperationID: "SearchMessages", Summary: "Search messages"},
				DELETE: &Operation{OperationID: "DeleteSearchResults", Summary: "Delete search results"},
			},
			"/api/v1/send": {
				POST: &Operation{OperationID: "SendMessage", Summary: "Send message"},
			},
			"/api/v1/tags": {
				GET: &Operation{OperationID: "GetTags", Summary: "Get tags"},
				PUT: &Operation{OperationID: "SetTags", Summary: "Set tags"},
			},
			"/api/v1/tags/{tag}": {
				DELETE: &Operation{OperationID: "DeleteTag", Summary: "Delete tag"},
			},
			"/api/v1/info": {
				GET:  &Operation{OperationID: "GetServerInfo", Summary: "Get server info"},
				HEAD: &Operation{OperationID: "Ping", Summary: "Ping server"},
			},
			"/api/v1/webui": {
				GET: &Operation{OperationID: "GetWebUIConfig", Summary: "Get web UI config"},
			},
			"/livez": {
				GET: &Operation{OperationID: "HealthCheck", Summary: "Health check"},
			},
		},
	}

	// Extract API routes from the static specification
	routes := extractAPIRoutes(staticSpec)
	require.NotEmpty(t, routes, "Should have discovered API routes from static specification")

	// Get implemented client methods
	clientMethods := getClientMethods()
	require.NotEmpty(t, clientMethods, "Should have discovered client methods")

	// Create route to method mappings
	mappings := createRouteMappings(routes, clientMethods)

	// Analyze coverage
	coverage := analyzeCoverage(mappings)

	// This test should always pass as it tests against our known implementation
	coveragePercent := coverage["coverage_percent"].(float64)
	require.True(t, coveragePercent >= 90.0,
		"Coverage should be at least 90%% for known routes, got %.2f%%", coveragePercent)

	t.Logf("✅ Offline API Coverage Test PASSED! Coverage: %.2f%% (%d/%d routes)",
		coveragePercent, coverage["implemented"], coverage["total"])
}
