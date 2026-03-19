# goenv

The most over-engineered environment variable reader in the history of Go programming. A full-blown, battle-tested, enterprise-grade package that reads ONE environment variable and tells you if you're in prod or dev.

## Why does this exist?

Because `os.Getenv("ENV")` is 16 characters and sometimes you just need a whole fucking package to do the same thing but with ✨style✨.

## Features

- Reads the `ENV` environment variable. That's it. That's the feature.
- Defaults to `prod` because production is where dreams go to die and that's where you probably are.
- Can tell you if you're in `dev`. Revolutionary.
- Zero dependencies. Not even `fmt`. We don't need that shit.
- 100% test coverage because we're professionals.

## Usage

```go
package main

import (
    "fmt"

    "github.com/psyb0t/goenv"
)

func main() {
    // Get the current environment
    env := goenv.Get() // "prod" or "dev"
    fmt.Println("running in", env)

    // Check if prod
    if goenv.IsProd() {
        fmt.Println("don't fuck this up")
    }

    // Check if dev
    if goenv.IsDev() {
        fmt.Println("break whatever you want")
    }
}
```

## Environment Variable

```bash
export ENV=dev   # you're developing
export ENV=prod  # you're in production
export ENV=      # also production, because paranoia is a feature
```

## API Reference

| Function | What it does | What it's basically doing |
|---|---|---|
| `Get()` | Returns `"prod"` or `"dev"` | `if os.Getenv("ENV") == "dev" { return "dev" } else { return "prod" }` |
| `IsProd()` | Returns `true` if prod | `Get() == "prod"` |
| `IsDev()` | Returns `true` if dev | `Get() == "dev"` |

## Constants

| Constant | Value | Surprise level |
|---|---|---|
| `EnvVarName` | `"ENV"` | None |
| `Prod` | `"prod"` | None |
| `Dev` | `"dev"` | None |

## FAQ

**Q: Do I really need a package for this?**
A: No. But here we are.

**Q: What if I set ENV to "staging"?**
A: You get prod. Because if it's not explicitly dev, it's prod. We don't trust you.

**Q: What about ENV=test?**
A: Also prod. See above.

**Q: Is this thread-safe?**
A: It calls `os.Getenv()`. Take that however you want.

**Q: Why not just use os.Getenv directly?**
A: Because then you'd have to remember the variable name, write a switch statement, handle the default case, and before you know it you've written this package anyway.

## License

MIT. Free as in "do whatever the hell you want with it". Which, given what this package does, isn't saying much.
