# Robinhood Metrics

Why not calculate Robinhood metrics just like you do for system metrics?

I find it a hassle to calculate ytd realized earnings on robinhood for options + stocks so I made this as a hobby. It's terrible and I bet bugs will arise.
However, I will do my best to address these issues...well at least until robinhood builds it in to their app!

Current Site is still a work in progress!. Feel free to ping me via filing github issues.

Navigate to `http://localhost:8080/` and login with your username, password and MFA.

You'll get redirected to a page which displays the following metrics

- Realized Earnings by Year
- Realized Earnings by Ticker
- Realized Earnings by type of transaction

# Local Development

```bash
export DEV=true

# To run site locally (localhost 8080)
go mod tidy
go run main.go

# To run linters
tools/trunk check
```

Navigate to `http://localhost:8080/` and login with your username, password and MFA.

I'm hosting this site for my own usability. if you'd like to run this locally, feel free to clone and run locally!
