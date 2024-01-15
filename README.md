# Robinhood Metrics

Why not calculate Robinhood metrics just like you do for system metrics?
Current Site is still a work in progress! Feel free to ping me via filing github issues.


# Local Development
```bash
export DEV=true

# To run site locally (localhost 8080)
go mod tidy
go run main.go

# To run linters
tools/trunk check
```

Navigate to `http://localhost:8080/`