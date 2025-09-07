# API Coverage Testing

This document explains the API coverage testing system that ensures the Mailpit Go client library implements all available Mailpit API endpoints.

## Overview

The API coverage test (`e2e_api_coverage_test.go`) automatically:

1. **Fetches** the latest Mailpit OpenAPI specification from the official repository
2. **Extracts** all API routes and their details
3. **Maps** routes to implemented client methods
4. **Reports** coverage statistics and missing implementations
5. **Fails** if required routes are missing or coverage falls below 95%

## Test Files

- `e2e_api_coverage_test.go` - Main API coverage test
- `scripts/api-coverage.sh` - Maintenance utilities script

## Running Tests

### Quick Test (Recommended)
```bash
# Run both online and offline coverage tests
make test-api-coverage

# Or run directly with Go
go test -v -run TestAPIRouteCoverage -timeout=3m
```

### Offline Test Only (Faster)
```bash
# When you don't need to fetch latest spec
go test -v -run TestAPIRouteCoverageOffline -timeout=30s
```

### Using the Maintenance Script
```bash
# Full test with utilities
./scripts/api-coverage.sh test

# Offline test only
./scripts/api-coverage.sh test-offline

# Check dependencies
./scripts/api-coverage.sh check-deps
```

## How It Works

### 1. Route Discovery
The test fetches the OpenAPI specification from:
- Primary: `https://raw.githubusercontent.com/axllent/mailpit/develop/server/ui/api/v1/swagger.json`
- Fallback: `https://raw.githubusercontent.com/axllent/mailpit/master/server/ui/api/v1/swagger.json`

### 2. Route Mapping
Routes are mapped to client methods using a predefined mapping table in `findMatchingMethod()`:

```go
routeMethodMap := map[string]string{
    "GET:/api/v1/messages": "ListMessages",
    "POST:/api/v1/send":    "SendMessage",
    // ... more mappings
}
```

### 3. Coverage Analysis
- **Required Routes**: Core API functionality that must be implemented
- **Optional Routes**: Advanced features that may not be available in all setups
- **Coverage Threshold**: 95% minimum coverage required

### 4. Quality Checks
The test also identifies potentially incorrect mappings and suggests improvements.

## Maintaining the Test

### When API Routes Change

1. **New Routes Added**: The test will detect and fail, showing missing routes
2. **Routes Modified**: Update the mapping table if parameter names change
3. **Routes Removed**: Remove from mapping table and optional routes list

### Adding New Route Mappings

1. **Identify the Route**: Check test output for missing routes
2. **Implement Client Method**: Add the method to the `Client` interface and implementation
3. **Update Mapping**: Add to `routeMethodMap` in `findMatchingMethod()`
4. **Test**: Run coverage test to verify

Example:
```go
// In findMatchingMethod(), add:
"POST:/api/v1/new-endpoint": "NewEndpointMethod",

// In client.go interface, add:
NewEndpointMethod(ctx context.Context, param string) (*Response, error)
```

### Marking Routes as Optional

Some routes may not be available in all Mailpit configurations. Mark them as optional:

```go
// In checkRequiredRoutes(), add to optionalRoutes:
"GET:/api/v1/optional-feature": true,
```

### Updating OpenAPI Specification URL

If the Mailpit repository changes the spec location:

```bash
./scripts/api-coverage.sh update-spec-url "https://new-url/swagger.json"
```

## Test Output Interpretation

### ✅ Successful Test
```
✅ API Route Coverage Test PASSED!
   Coverage: 100.00% (23/23 routes implemented)
```

### ❌ Failed Test - Missing Routes
```
❌ MISSING ROUTES:
  GET /api/v1/new-endpoint
    Summary: New endpoint description
    Operation ID: NewEndpointParams
    Notes: No matching client method found
```

### ⚠️ Quality Warnings
```
⚠️ MAPPING QUALITY WARNINGS:
  - PUT /api/v1/tags/{Tag} is mapped to DeleteTag() but should probably be RenameTag()
  Consider implementing proper methods for these routes.
```

## Route Categories

### Core Message Operations
- List, get, delete messages
- Message headers, source, parts, attachments
- Message validation (HTML, links, spam)

### Search Operations
- Search messages with queries
- Delete search results

### Send Operations
- Send messages via SMTP

### Tags Operations
- Get, set, delete tags
- Tag message associations

### Server Operations
- Server info, health checks
- Web UI configuration

### Chaos Testing
- Chaos engineering configuration (optional)

## Best Practices

### 1. Regular Testing
- Run API coverage tests in CI/CD pipeline
- Test against multiple Mailpit versions
- Monitor for API changes in Mailpit releases

### 2. Implementation Priority
1. **Required Routes**: Implement immediately when missing
2. **Optional Routes**: Implement based on user needs
3. **Quality Issues**: Address mapping problems for better accuracy

### 3. Documentation
- Update route mappings when implementing new methods
- Document any special handling or limitations
- Keep API coverage documentation current

### 4. Error Handling
- Handle optional endpoints gracefully (404 errors expected)
- Provide clear error messages for missing implementations
- Test error scenarios as well as success cases

## Integration with CI/CD

Add to your CI pipeline:

```yaml
# GitHub Actions example
- name: Test API Coverage
  run: |
    go test -v -run TestAPIRouteCoverage -timeout=3m
    if [ $? -ne 0 ]; then
      echo "API coverage test failed. Check for missing route implementations."
      exit 1
    fi
```

## Troubleshooting

### Common Issues

1. **Network Failures**: Test falls back to offline mode automatically
2. **Spec Format Changes**: Update parsing logic in `extractAPIRoutes()`
3. **Parameter Name Variations**: Add alternative mappings in `normalizePathParameters()`
4. **New Endpoint Types**: Extend route categorization logic

### Debug Mode

For detailed debugging, modify the test to log more information:

```go
// Add more verbose logging
t.Logf("Route: %+v", route)
t.Logf("Mapping attempt: %s", routeKey)
```

### Manual Spec Inspection

Download and examine the spec manually:

```bash
curl -s https://raw.githubusercontent.com/axllent/mailpit/develop/server/ui/api/v1/swagger.json | jq '.'
```

## Future Enhancements

Planned improvements:

1. **Automatic Stub Generation**: Generate method stubs for missing routes
2. **Version Compatibility**: Test against multiple Mailpit API versions
3. **Performance Metrics**: Track API response times and performance
4. **Schema Validation**: Validate request/response schemas match spec
5. **Integration Tests**: Combine with E2E tests for full validation

## Contributing

When contributing new API methods:

1. Check the coverage test first to understand missing routes
2. Implement the client method following existing patterns
3. Update the route mapping table
4. Run the coverage test to verify
5. Update documentation if needed

This ensures the library maintains comprehensive API coverage and stays current with Mailpit development.
