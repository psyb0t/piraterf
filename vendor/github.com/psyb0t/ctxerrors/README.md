# ctxerrors

```
 ####  ##### #    # ###### #####  #####   ####  #####   ####  
#    #   #    #  #  #      #    # #    # #    # #    # #      
#        #     ##   #####  #    # #    # #    # #    #  ####  
#        #     ##   #      #####  #####  #    # #####       # 
#    #   #    #  #  #      #   #  #   #  #    # #   #  #    # 
 ####    #   #    # ###### #    # #    #  ####  #    #  ####  
```

fuck yeah, another Go error handling package 🖕

A Go library that wraps errors with context information (file, line, function) because debugging without context is like trying to find your dick in the dark.

## Table of Contents

- [Installation](#installation)
- [What the fuck does it do?](#what-the-fuck-does-it-do)
  - [Functions](#functions)
- [Usage](#usage)
  - [Creating new errors](#creating-new-errors)
  - [Wrapping existing errors](#wrapping-existing-errors)
  - [Formatted wrapping](#formatted-wrapping)
- [Error output](#error-output)
  - [Error chaining](#error-chaining)
  - [Stupid inline chaining](#stupid-inline-chaining)
  - [Unwrapping errors](#unwrapping-errors)
- [More stupid fucking examples](#more-stupid-fucking-examples)
  - [Annoyingly complex tangled bullshit](#annoyingly-complex-tangled-bullshit)
  - [Ridiculously stupid chain of doom](#ridiculously-stupid-chain-of-doom)
- [License](#license)
- [Why?](#why)

## Installation

```bash
go get github.com/psyb0t/ctxerrors
```

## What the fuck does it do?

This package automatically captures where your errors happen in your code. No more hunting through logs like some dickless detective wondering where the fuck that error came from.

### Functions

- **New()** - Creates a new error with location context
- **Wrap()** - Wraps existing errors with additional context and location
- **Wrapf()** - Like Wrap() but with printf-style formatting because we're not animals

All functions return an `*ErrorWithContext` that implements the standard `error` interface and supports `errors.Unwrap()`, `errors.Is()`, and `errors.As()` because Go's error handling conventions aren't completely ass-backwards.

## Usage

### Creating new errors

```go
import "github.com/psyb0t/ctxerrors"

func doSomething() error {
    return ctxerrors.New("shit went sideways")
}
```

### Wrapping existing errors

```go
func processFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return ctxerrors.Wrap(err, "failed to open file")
    }
    defer file.Close()
    
    // do stuff...
    
    return nil
}
```

### Formatted wrapping

```go
func connectToDatabase(host string, port int) error {
    conn, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d", host, port))
    if err != nil {
        return ctxerrors.Wrapf(err, "failed to connect to database at %s:%d", host, port)
    }
    defer conn.Close()
    
    return nil
}
```

## Error output

When shit hits the fan, you get detailed context:

```
failed to connect to database at localhost:5432: dial tcp [::1]:5432: connect: connection refused [/path/to/your/file.go:42 in main.connectToDatabase]
```

### Error chaining

When you wrap ctxerrors in a chain, each layer shows its context:

```go
func readConfig() error {
    return ctxerrors.New("config file missing")
}

func initDatabase() error {
    if err := readConfig(); err != nil {
        return ctxerrors.Wrap(err, "failed to read config")
    }
    return nil
}

func startServer() error {
    if err := initDatabase(); err != nil {
        return ctxerrors.Wrap(err, "database initialization failed")
    }
    return nil
}
```

Output shows the full chain with context from each wrap:

```
database initialization failed: failed to read config: 
config file missing [/path/to/server.go:12 in main.readConfig] 
[/path/to/server.go:17 in main.initDatabase] 
[/path/to/server.go:24 in main.startServer]
```

### Stupid inline chaining

Or if you're a masochist and like one-liners:

```go
func clusterfuck() error {
    return ctxerrors.Wrap(
        ctxerrors.Wrap(
            ctxerrors.New("original fuckup"),
            "second layer of shit"),
        "final layer of despair")
}
```

Output:

```
final layer of despair: second layer of shit: original fuckup 
[/path/to/file.go:42 in main.clusterfuck] 
[/path/to/file.go:41 in main.clusterfuck] 
[/path/to/file.go:40 in main.clusterfuck]
```

### Unwrapping errors

You can unwrap the chain to get to the original error:

```go
err := startServer()
originalErr := errors.Unwrap(errors.Unwrap(err))
// originalErr is now the "config file missing" error

// Or use errors.Is() to check if specific error is in the chain
if errors.Is(err, someSpecificError) {
    // handle it
}

// Or use errors.As() to get the ErrorWithContext
var ctxErr *ctxerrors.ErrorWithContext
if errors.As(err, &ctxErr) {
    fmt.Printf("Error occurred in %s at line %d\n", ctxErr.funcName, ctxErr.line)
}
```

No more guessing where the fuck everything went tits up.

## More stupid fucking examples

### Annoyingly complex tangled bullshit

Because sometimes you write code like a fucking maniac:

```go
func processUserShitWithStupidNesting(userID int) error {
    validateUser := func(id int) error {
        if id <= 0 {
            return ctxerrors.New("invalid user ID: must be greater than zero")
        }
        return nil
    }
    
    fetchUserData := func(id int) error {
        if rand.Intn(3) == 0 {
            return ctxerrors.Wrapf(
                errors.New("connection timeout"),
                "failed to fetch user data for user ID %d", id)
        }
        return nil
    }
    
    processPermissions := func(id int) error {
        checkAdminRights := func() error {
            if rand.Intn(2) == 0 {
                return ctxerrors.New("user lacks admin privileges")
            }
            return nil
        }
        
        if err := checkAdminRights(); err != nil {
            return ctxerrors.Wrapf(err, "permission check failed for user %d", id)
        }
        return nil
    }
    
    // Chain all this shit together
    if err := validateUser(userID); err != nil {
        return ctxerrors.Wrap(err, "user validation step failed")
    }
    
    if err := fetchUserData(userID); err != nil {
        return ctxerrors.Wrap(err, "data fetching step failed")
    }
    
    if err := processPermissions(userID); err != nil {
        return ctxerrors.Wrap(err, "permission processing step failed")
    }
    
    return nil
}
```

When this clusterfuck fails, you get a beautiful trace:

```
data fetching step failed: failed to fetch user data for user ID 42: 
connection timeout [main.go:15 in processUserShitWithStupidNesting.func2] 
[main.go:35 in processUserShitWithStupidNesting]
```

### Ridiculously stupid chain of doom

For when you really want to piss off your future self:

```go
func performStupidlyComplexOperation() error {
    // Because apparently we hate ourselves and everyone who reads this code
    return func() error {
        // Welcome to nested function hell, population: you
        if err := func() error {
            // This is where sanity comes to die
            if err := func() error {
                // At this point we're just fucking with people
                if err := func() error {
                    // The beginning of the end
                    if err := func() error {
                        // Rock bottom of this shitshow
                        return ctxerrors.New("step 1 went to shit")
                    }(); err != nil {
                        // Step 2: electric boogaloo of failure
                        return ctxerrors.Wrap(err, "step 2 couldn't handle step 1's bullshit")
                    }
                    // If we somehow made it this far, we're lying
                    return nil
                }(); err != nil {
                    // Step 3: the reckoning
                    return ctxerrors.Wrap(err, "step 3 is having a mental breakdown")
                }
                // Still pretending everything is fine
                return nil
            }(); err != nil {
                // Step 4: fuck it, we're done trying
                return ctxerrors.Wrap(err, "step 4 said fuck this shit")
            }
            // The calm before the storm
            return nil
        }(); err != nil {
            // The final boss of this clusterfuck
            return ctxerrors.Wrap(err, "the entire fucking operation is fucked")
        }
        // Narrator: it was not fine
        return nil
    }() // Because we needed one more layer of stupid
}
```

Output when everything goes to hell:

```
the entire fucking operation is fucked: step 4 said fuck this shit: 
step 3 is having a mental breakdown: step 2 couldn't handle step 1's bullshit: 
step 1 went to shit [main.go:42 in step1] [main.go:47 in step2] 
[main.go:53 in step3] [main.go:59 in step4] [main.go:65 in performStupidlyComplexOperation]
```

This shit makes debugging actually bearable instead of wanting to throw your laptop out the fucking window.

## License

MIT License - because lawyers are expensive and I don't want to deal with that shit

## Why?

Because Go's error handling is verbose as fuck and debugging without context is like trying to find your asshole with both hands tied behind your back. This shit makes it slightly less painful.