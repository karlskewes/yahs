# yahs

Yet another HTTP server.

Example implementations of common requirements for website serving applications.

## Requirements

1. Basic HTTP server - https://github.com/karlskewes/yahs/pull/1
1. GitHub Action CI to lint & test - https://github.com/karlskewes/yahs/pull/2
1. Graceful shutdown - https://github.com/karlskewes/yahs/pull/4
1. Testable application invocations. Split `main()` with `Run()` - https://github.com/karlskewes/yahs/pull/5
1. Enable importing into other applications, move `package main` to `cmd/..` - https://github.com/karlskewes/yahs/pull/6
1. Create an `App` type to hold config & state - https://github.com/karlskewes/yahs/pull/7
1. Functional Options pattern for config, support custom `http.Server{}` - https://github.com/karlskewes/yahs/pull/8
1. Rename package to `yahs` for simpler imports - https://github.com/karlskewes/yahs/pull/9
1. Rename HTTP Server from App to Server - https://github.com/karlskewes/yahs/pull/10
1. Custom Route handling - https://github.com/karlskewes/yahs/pull/11
1. Embedded filesystem for assets and HTML templates - https://github.com/karlskewes/yahs/pull/12
1. HTTP Route precedence - https://github.com/karlskewes/yahs/pull/13
1. Exit if unable to listen on address - https://github.com/karlskewes/yahs/pull/15
