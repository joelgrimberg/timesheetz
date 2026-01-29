# Timesheetz - Claude Instructions

## Workflow

Before implementing any feature or change:
1. Rephrase the request to confirm understanding
2. Wait for user approval before proceeding
3. Only implement after explicit confirmation

## Testing

After implementing changes:
1. Check if tests need to be added, modified, or removed
2. Update tests to match the new behavior
3. Run `go test ./...` to verify all tests pass
4. Fix any failing tests before considering the work complete
