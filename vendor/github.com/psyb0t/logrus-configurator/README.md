# logrus-configurator 🤖

Welcome to `logrus-configurator`, the badass sidekick for your logging adventures with Go! You've stumbled upon the dark alley of loggers where things get configured so slick that even your professors can't help but nod in approval. 😎

## What's This Shit About? 💩

`logrus-configurator` is a Go package that whips your `logrus` logger into shape without you breaking a sweat. Think of it as that one plugin at a rave that just knows how to tune things up. Want to set the log level? Bam! 🎚️ Prefer JSON over plain text? Wham! 📄 Want to know who called the logger? Boom! 🔍 It's got you covered.

## Features

- **No-nonsense log level setting** (trace your bugs or go full-on panic mode, we don't judge).
- **Formatting logs like a boss** with JSON or text formats – keep it structured or keep it simple.
- **Caller reporting** for when you need to backtrack who messed up. It's like `CSI` for your code.
- **Automated configuration** using environment variables, because who has time for manual setup?
- **Hook management API** for when you need custom logging destinations and advanced control.

## Usage Example

Ready to rock with `logrus-configurator`? Check this out.

### main.go

```go
package main

import (
	_ "github.com/psyb0t/logrus-configurator"
	"github.com/sirupsen/logrus"
)

func main() {
	// Here's where the magic happens - just logging some stuff.
	logrus.Trace("this shit's a trace") // Ninja mode, won't show unless you want it to.
	logrus.Debug("this shit's a debug") // Debugging like a boss.
	logrus.Info("this shit's an info")  // Cool, calm, and collected info.
	logrus.Warn("this shit's a warn")   // Warning: badass logger at work.
	logrus.Error("this shit's an error") // Oh crap, something went sideways.
	logrus.Fatal("this shit's a fatal")  // Critical hit! It's super effective!
}
```

### Crank It Up

Get your environment dialed in like the soundboard at a goth concert:

```bash
export LOG_LEVEL="trace"   # Choose the verbosity level.
export LOG_FORMAT="text"   # Pick your poison: json or text.
export LOG_CALLER="true"   # Decide if you want to see who's calling the logs.
```

Unleash the beast with:

```bash
go run main.go
```

And let the good times roll with the output:

```plaintext
time="2025-09-07T10:56:28Z" level=debug msg="logrus-configurator: level: trace, format: text, reportCaller: true" func="github.com/psyb0t/logrus-configurator.config.log()" file="log.go:30"
time="2025-09-07T10:56:28Z" level=trace msg="this shit's a trace" func="main.main()" file="main.go:9"
time="2025-09-07T10:56:28Z" level=debug msg="this shit's a debug" func="main.main()" file="main.go:10"
time="2025-09-07T10:56:28Z" level=info msg="this shit's an info" func="main.main()" file="main.go:11"
time="2025-09-07T10:56:28Z" level=warning msg="this shit's a warn" func="main.main()" file="main.go:12"
time="2025-09-07T10:56:28Z" level=error msg="this shit's an error" func="main.main()" file="main.go:13"
time="2025-09-07T10:56:28Z" level=fatal msg="this shit's a fatal" func="main.main()" file="main.go:14"
exit status 1
```

Wanna switch it up? Change the environment variables to mix the brew.

```bash
export LOG_LEVEL="warn"
export LOG_FORMAT="json"
export LOG_CALLER="false"
```

Then let it simmer with:

```bash
go run main.go
```

And enjoy the sweet sound of (almost) silence:

```plaintext
{"level":"warning","msg":"this shit's a warn","time":"2025-09-07T10:56:33Z"}
{"level":"error","msg":"this shit's an error","time":"2025-09-07T10:56:33Z"}
{"level":"fatal","msg":"this shit's a fatal","time":"2025-09-07T10:56:33Z"}
exit status 1
```

Whether you're in for a riot or a silent disco, `logrus-configurator` is your ticket. 🎟️ (check out all of the supported levels in [`level.go`](level.go))

## Advanced Hook Management 🚀

Need more control over your logging destinations? Here's some badass functions for managing custom hooks:

```go
import "github.com/psyb0t/logrus-configurator"

// Replace all hooks with your custom ones - go full nuclear
logrusconfigurator.SetHooks(myDbHook, mySlackHook, myCustomHook)

// Add a single hook without fucking up the existing setup
logrusconfigurator.AddHook(myMetricsHook)
```

Perfect for when you need to:
- Send errors to external monitoring systems
- Log to databases or message queues  
- Route different log levels to different destinations
- Build complex logging pipelines

The package handles stderr/stdout separation automatically, so your custom hooks play nice with the default console output.

## Testing & Quality 🧪

This package is tested harder than a Nokia 3310. This shit's got **96.3% test coverage** because nobody fucks around with quality here.

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage check (fails if below 90%)
make test-coverage

# Run linting
make lint

# Run everything (dependencies, linting, tests)
make all
```

### What Gets Tested

**Core Shit:**
- ✅ Environment variable parsing and configuration
- ✅ All log levels (trace, debug, info, warn, error, fatal, panic)
- ✅ JSON and text formatting with caller prettification
- ✅ Hook management and custom writers
- ✅ Error handling for fucked up configurations

**Advanced Shit:**
- ✅ `SetHooks()` and `AddHook()` API functions
- ✅ Caller prettification with complex function signatures
- ✅ Custom hook writers and level filtering
- ✅ Configuration edge cases and error scenarios

**Testing Philosophy:**
- Table-driven tests for consistency and maintainability
- Proper `require` vs `assert` patterns for better debugging
- 90% minimum coverage enforced in CI/CD
- Comprehensive edge case coverage

And that's damn it. You've just pimped your logger with military-grade reliability!

## Contribute

Got an idea? Throw in a PR! Found a bug? Raise an issue! Let's make `logrus-configurator` as tight as your favorite jeans.

## License

It's MIT. Free as in 'do whatever the hell you want with it', just don't blame me if shit hits the fan.