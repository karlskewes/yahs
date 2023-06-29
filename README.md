# go-yahs

Yet another HTTP server.

Example implementations of common requirements for website serving applications.

## Requirements

1. Basic HTTP server - https://github.com/karlskewes/go-yahs/pull/1
1. GitHub Action CI to lint & test - https://github.com/karlskewes/go-yahs/pull/2
1. Graceful shutdown - https://github.com/karlskewes/go-yahs/pull/4
1. Testable application invocations. Split `main()` with `Run()` - https://github.com/karlskewes/go-yahs/pull/5
1. Enable importing into other applications, move `package main` to `cmd/..` - https://github.com/karlskewes/go-yahs/pull/6
1. Create an `App` type to hold config & state - https://github.com/karlskewes/go-yahs/pull/7
