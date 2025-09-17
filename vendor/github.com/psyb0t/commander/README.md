```
  ☩ ═════════════════════════════════════════════════════════ ☩
     __________  __  _____  ______    _   ______  __________ 
    / ____/ __ \/  |/  /  |/  /   |  / | / / __ \/ ____/ __ \
   / /   / / / / /|_/ / /|_/ / /| | /  |/ / / / / __/ / /_/ /
  / /___/ /_/ / /  / / /  / / ___ |/ /|  / /_/ / /___/ _, _/ 
  \____/\____/_/  /_/_/  /_/_/  |_/_/ |_/_____/_____/_/ |_|
                             🐚💀🔥
  ☩ ═════════════════════════════════════════════════════════ ☩
```

# Commander 🐚

## SOMEBODY STOP ME! 💚 Command execution from hell's kitchen 🔥😈

Commander takes Go's `os/exec` and transforms it from a fucking disaster 💥 into something that actually works - P-A-R-T-Y! 🎉 This shit wraps all the garbage 🗑️ that makes you want to violate everything holy and spawn some digital violence 🔪💻. No more hanging processes (they picked the wrong god to pray to! ⚰️), no more race conditions giving you digital hemorrhoids 🩸, no more timeout bullshit ⏰💩 that makes you question why you didn't just become a tortured soul in the first place 👹.

**SMOKIN'! 🚬**: Stream output like the green-faced monster 👹💚 you were born to be, terminate processes with malevolent glee 💀🔥, mock everything without losing your goddamn sanity 🧠💥, and handle errors like a true hellspawn console warrior 👺💻. I am returned from the darkness to tear your shitty command execution apart! ⚔️🖥️🔥

## Installation

```bash
go get github.com/psyb0t/commander
```

## The Interfaces 🛠️ (Know Your Weapons 🗡️⚔️)

### Commander Interface 💚🎭 - The main event (SOMEBODY STOP ME! 👹)
```go
type Commander interface {
    // Fire and forget - P-A-R-T-Y! 🔥🎉 Just run the damn thing 💨
    Run(ctx context.Context, name string, args []string, opts ...Option) error
    
    // Get stdout/stderr separated - it's showtime! 🎭💥
    Output(ctx context.Context, name string, args []string, opts ...Option) (stdout []byte, stderr []byte, err error)
    
    // Mix that output together - VIOLATE EVERYTHING HOLY! 👹💀
    CombinedOutput(ctx context.Context, name string, args []string, opts ...Option) (output []byte, err error)
    
    // Start a process you can control - somebody stop me! 🛑👹🔥
    Start(ctx context.Context, name string, args []string, opts ...Option) (Process, error)
}
```

### Process Interface 🔥👹 - Hellspawn fuckery from the abyss ⚰️💀
```go
type Process interface {
    Start() error                                            // Spawn the beast - it's PARTY TIME! 🎉👹🔥
    Wait() error                                            // Wait for the carnage to finish ⏰💀
    StdinPipe() (io.WriteCloser, error)                    // Feed the machine - VIOLATE EVERYTHING! 🚰🔪
    Stream(stdout, stderr chan<- string)                    // Stream the chaos live - witness the violence! 🌍💻📡⚡
    Stop(ctx context.Context) error                         // They picked the wrong god to pray to! ⚰️👹💀
    Kill(ctx context.Context) error                        // Somebody stop me from this beautiful murder! 🔫💥💚
    PID() int                                               // Get the process ID - know thy enemy! 🎯👹🔢
}
```

### Options ⚙️👹 - Pimp your malevolent machinery 🔥💚

**Command Execution Options:**
```go
func WithStdin(stdin io.Reader) Option        // Feed the beast - it's party time! 🍽️👹🎉
func WithEnv(env []string) Option            // Corrupt the environment - spawn the chaos! 🌍💻🔥👺
func WithDir(dir string) Option              // Choose your battlefield - violate everything holy! 📁🗂️⚰️
```


## Basic Usage 💚⚔️ - The fundamentals of digital violence 🔥💀

### Simple Command Execution
```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"

    "github.com/psyb0t/commander"
    commonerrors "github.com/psyb0t/common-go/errors"
)

func main() {
    cmd := commander.New()
    ctx := context.Background()

    // Just run dat shit and forget about it - wicked!
    err := cmd.Run(ctx, "echo", []string{"hello world"})
    if err != nil {
        log.Fatal("Failed to run command - what a fucking disaster:", err)
    }

    // Get da output like a civilized person, innit
    stdout, stderr, err := cmd.Output(ctx, "ls", []string{"-la", "/tmp"})
    if err != nil {
        log.Fatal("Command failed - dis is well fucked:", err)
    }
    
    fmt.Printf("Files:\n%s\n", stdout)
    if len(stderr) > 0 {
        fmt.Printf("Errors (oh for fuck's sake):\n%s\n", stderr)
    }

    // When you don't give a toss about separating streams
    output, err := cmd.CombinedOutput(ctx, "git", []string{"status"})
    if err != nil {
        log.Fatal("Git failed - typical fucking git:", err)
    }
    fmt.Printf("Git says (probably some bullshit):\n%s\n", output)
}
```

## Advanced Shit - Real-time Streaming, bruv

Want to see what's happening while it's happening? Here's how you stream dat shit live - it's well good:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/psyb0t/commander"
)

func main() {
    cmd := commander.New()
    ctx := context.Background()

    // Start a long-running process
    proc, err := cmd.Start(ctx, "ping", []string{"-c", "10", "google.com"})
    if err != nil {
        log.Fatal("Failed to start ping:", err)
    }

    // Create channels for live streaming
    stdout := make(chan string, 100)  // Buffer it so we don't block
    stderr := make(chan string, 100)

    // Start streaming (this is non-blocking)
    proc.Stream(stdout, stderr)

    // Read the streams as they come in
    go func() {
        for line := range stdout {
            fmt.Printf("[PING] %s\n", line)
        }
    }()

    go func() {
        for line := range stderr {
            fmt.Printf("[ERROR] %s\n", line)
        }
    }()

    // Wait for the process to finish
    err = proc.Wait()
    if err != nil {
        fmt.Printf("Ping finished with error: %v\n", err)
    } else {
        fmt.Println("Ping completed successfully!")
    }
}
```

## Process Control - Be the Boss

### Graceful Termination
```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "time"

    "github.com/psyb0t/commander"
    commonerrors "github.com/psyb0t/common-go/errors"
)

func main() {
    cmd := commander.New()
    ctx := context.Background()

    // Start something that runs forever
    proc, err := cmd.Start(ctx, "tail", []string{"-f", "/var/log/syslog"})
    if err != nil {
        log.Fatal("Failed to start tail:", err)
    }

    // Let it run for a bit
    time.Sleep(2 * time.Second)

    // Now shut it down gracefully (SIGTERM first, SIGKILL after 5 seconds if needed)
    fmt.Println("Shutting down gracefully...")
    stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    err = proc.Stop(stopCtx)
    
    if err == nil {
        fmt.Println("Process stopped cleanly")
    } else if errors.Is(err, commonerrors.ErrTerminated) {
        fmt.Println("Process terminated gracefully (SIGTERM)")
    } else if errors.Is(err, commonerrors.ErrKilled) {
        fmt.Println("Process had to be killed (SIGKILL)")
    } else {
        fmt.Printf("Stop failed: %v\n", err)
    }
}
```

### Spank you! Spank you very much! (Immediate Process Termination)
```go
func justKillIt() {
    cmd := commander.New()
    ctx := context.Background()

    proc, _ := cmd.Start(ctx, "sleep", []string{"1000"})
    
    // Like a glove! No mercy, just beautiful violence
    err := proc.Kill(ctx)
    if errors.Is(err, commonerrors.ErrKilled) {
        fmt.Println("Process killed with SIGKILL - somebody stop me!")
    } else if err != nil {
        fmt.Printf("Kill failed: %v\n", err)
    }
}
```

## Custom Kill Signals 💀⚔️ - Choose your weapon of destruction

### Using Custom Signals for Graceful Termination
```go
package main

import (
    "context"
    "fmt"
    "syscall"
    "time"

    "github.com/psyb0t/commander"
)

func killWithStyle() {
    cmd := commander.New()
    ctx := context.Background()

    proc, err := cmd.Start(ctx, "your-daemon", []string{"--config", "prod.yml"})
    if err != nil {
        panic(err)
    }

    // Give it 10 seconds to shut down gracefully with SIGINT instead of SIGTERM
    stopCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    
    err = proc.Stop(stopCtx)
    if err != nil {
        fmt.Printf("Process stopped with: %v\n", err)
    }
}

func useUserSignals() {
    cmd := commander.New()
    ctx := context.Background()

    proc, err := cmd.Start(ctx, "nginx", []string{"-g", "daemon off;"})
    if err != nil {
        panic(err)
    }

    // Nginx responds to SIGUSR1 for graceful reload
    stopCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    err = proc.Stop(stopCtx)
    if err != nil {
        fmt.Printf("Nginx graceful reload result: %v\n", err)
    }
}
```

## Timeout Handling - Patience is for suckers

### Context-Based Timeouts (The Right Way™️)
```go
func contextTimeout() {
    cmd := commander.New()
    
    // This will timeout after 2 seconds - context controls everything!
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    err := cmd.Run(ctx, "sleep", []string{"10"})
    if errors.Is(err, commonerrors.ErrTimeout) {
        fmt.Println("Bingo! Command timed out like a champ!")
    }
}

// Stop with custom timeout - no more redundant bullshit!
func stopWithTimeout() {
    cmd := commander.New()
    ctx := context.Background()
    
    proc, _ := cmd.Start(ctx, "sleep", []string{"100"})
    
    // Give it 3 seconds to die gracefully, then SIGKILL the fucker
    stopCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()
    
    err := proc.Stop(stopCtx) // Clean as fuck - context controls timeout!
    if errors.Is(err, commonerrors.ErrTerminated) {
        fmt.Println("Process gracefully terminated!")
    } else if errors.Is(err, commonerrors.ErrKilled) {
        fmt.Println("Process was force killed after timeout!")
    }
}
```


## API Migration Guide 🔄 - From the old shit to the new hotness

**Old API (Redundant bullshit):**
```go
// OLD - Don't use this crap anymore!
err := proc.Stop(ctx, 5*time.Second) // WTF? Both ctx AND timeout?
```

**New API (Clean as fuck):**
```go
// NEW - Context controls everything like a boss!
stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
err := proc.Stop(stopCtx) // One source of truth for timeouts

// Graceful stop with SIGTERM then SIGKILL
err := proc.Stop(stopCtx)

// No timeout? No problem - immediate force kill
err := proc.Stop(context.Background()) // No deadline = force kill
```

**Why the change?** Because having both `ctx` and `timeout` parameters was fucking redundant! Now the user has full control - use `context.WithTimeout()`, `context.WithDeadline()`, `context.WithCancel()`, or any other context pattern. Much cleaner and follows Go idioms properly.

## Input and Environment - Feeding time at the process zoo

### Stdin Input
```go
func stdinExample() {
    cmd := commander.New()
    ctx := context.Background()

    // Feed some data to wc to count lines
    input := strings.NewReader("line 1\nline 2\nline 3\nline 4\n")
    
    stdout, _, err := cmd.Output(ctx, "wc", []string{"-l"}, 
        commander.WithStdin(input))
    
    if err != nil {
        log.Fatal("wc failed:", err)
    }

    fmt.Printf("Line count: %s", stdout) // Should print "4" - Smokin'!
}
```

### Environment Variables
```go
func environmentExample() {
    cmd := commander.New()
    ctx := context.Background()

    // Set custom environment
    stdout, _, err := cmd.Output(ctx, "sh", []string{"-c", "echo $CUSTOM_VAR $ANOTHER_VAR"}, 
        commander.WithEnv([]string{
            "CUSTOM_VAR=hello",
            "ANOTHER_VAR=world",
        }))
    
    if err != nil {
        log.Fatal("Shell command failed:", err)
    }

    fmt.Printf("Environment output: %s", stdout) // Should print "hello world" - Alllllrighty then!
}
```

### Working Directory
```go
func workingDirectoryExample() {
    cmd := commander.New()
    ctx := context.Background()

    // Run pwd in /tmp
    stdout, _, err := cmd.Output(ctx, "pwd", nil, 
        commander.WithDir("/tmp"))
    
    if err != nil {
        log.Fatal("pwd failed:", err)
    }

    fmt.Printf("Current directory: %s", stdout) // Should print "/tmp"
}
```

### All Options Combined
```go
func kitchenSinkExample() {
    cmd := commander.New()
    ctx := context.Background()

    input := strings.NewReader("some input data")
    
    // Use context timeout instead of WithTimeout option
    timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    stdout, stderr, err := cmd.Output(timeoutCtx, "cat", nil,
        commander.WithStdin(input),
        commander.WithDir("/tmp"),
        commander.WithEnv([]string{"LANG=en_US.UTF-8"}))
    
    if err != nil {
        log.Fatal("Kitchen sink failed:", err)
    }

    fmt.Printf("Output: %s\n", stdout)
    if len(stderr) > 0 {
        fmt.Printf("Errors: %s\n", stderr)
    }
}
```

## Advanced Streaming - Multiple Listeners

You can have multiple channels listening to the same process output:

```go
func multipleListeners() {
    cmd := commander.New()
    ctx := context.Background()

    proc, err := cmd.Start(ctx, "ping", []string{"-c", "5", "google.com"})
    if err != nil {
        log.Fatal("Failed to start ping:", err)
    }

    // Create multiple listeners
    logger := make(chan string, 100)
    display := make(chan string, 100)
    storage := make(chan string, 100)

    // All three will get the same data
    proc.Stream(logger, nil)
    proc.Stream(display, nil) 
    proc.Stream(storage, nil)

    // Handle each stream differently
    go func() {
        for line := range logger {
            log.Printf("[LOG] %s", line)
        }
    }()

    go func() {
        for line := range display {
            fmt.Printf("[DISPLAY] %s\n", line)
        }
    }()

    var stored []string
    go func() {
        for line := range storage {
            stored = append(stored, line)
        }
        fmt.Printf("Stored %d lines total\n", len(stored))
    }()

    err = proc.Wait()
    if err != nil {
        fmt.Printf("Process failed: %v\n", err)
    }
}
```

## Concurrent Execution - Go Wild

Run multiple commands at the same time like a fucking machine:

```go
func concurrentExecution() {
    cmd := commander.New()
    ctx := context.Background()

    // Commands to run concurrently
    commands := []struct {
        name string
        args []string
    }{
        {"echo", []string{"first"}},
        {"echo", []string{"second"}},
        {"echo", []string{"third"}},
        {"sleep", []string{"1"}},
        {"date", nil},
    }

    var wg sync.WaitGroup
    results := make(chan string, len(commands))

    // Launch all commands concurrently
    for _, cmdInfo := range commands {
        wg.Add(1)
        go func(name string, args []string) {
            defer wg.Done()
            
            stdout, _, err := cmd.Output(ctx, name, args)
            if err != nil {
                results <- fmt.Sprintf("ERROR: %s %v failed: %v", name, args, err)
                return
            }
            
            results <- fmt.Sprintf("SUCCESS: %s %v -> %s", name, args, strings.TrimSpace(string(stdout)))
        }(cmdInfo.name, cmdInfo.args)
    }

    // Wait for all to complete
    wg.Wait()
    close(results)

    // Show results
    fmt.Println("Concurrent execution results:")
    for result := range results {
        fmt.Printf("  %s\n", result)
    }
}
```

## Error Handling - Know What Went Wrong

The package gives you specific error types so you know exactly what happened:

```go
func errorHandling() {
    cmd := commander.New()
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    proc, err := cmd.Start(ctx, "sleep", []string{"10"})
    if err != nil {
        log.Fatal("Failed to start process:", err)
    }

    err = proc.Wait()
    if err == nil {
        fmt.Println("✅ Process completed successfully")
    } else if errors.Is(err, commonerrors.ErrTimeout) {
        fmt.Println("❌ Process timed out")
    } else if errors.Is(err, commonerrors.ErrTerminated) {
        fmt.Println("⚠️  Process was terminated (SIGTERM)")
    } else if errors.Is(err, commonerrors.ErrKilled) {
        fmt.Println("💀 Process was killed (SIGKILL)")
    } else {
        fmt.Printf("💥 Process failed: %v\n", err)
    }
}
```

## Testing - Mock That Shit

The package comes with a comprehensive mocking system so you can test without actually running commands:

### Basic Mocking
```go
func TestMyFunction(t *testing.T) {
    mock := commander.NewMock()
    defer func() {
        if err := mock.VerifyExpectations(); err != nil {
            t.Error("Mock expectations failed:", err)
        }
    }()

    // Set up expectations
    mock.Expect("git", "status").ReturnOutput([]byte("On branch main\nnothing to commit, working tree clean"))
    mock.Expect("git", "push").ReturnError(errors.New("push failed"))

    // Use the mock in your code (it implements Commander interface)
    err := myDeployFunction(mock)
    
    // Your function should handle the push failure gracefully
    assert.Error(t, err)
}
```

### Advanced Argument Matching
```go
func TestWithMatchers(t *testing.T) {
    mock := commander.NewMock()
    defer mock.VerifyExpectations()

    // Exact matching (default)
    mock.Expect("echo", "hello").ReturnOutput([]byte("hello"))

    // Regex matching
    mock.ExpectWithMatchers("grep", 
        commander.Regex("^error.*"),  // First arg must match regex
        commander.Exact("logfile.txt"))  // Second arg must be exact

    // Wildcard matching
    mock.ExpectWithMatchers("find", commander.Any(), commander.Any())

    // Mixed matching
    mock.ExpectWithMatchers("rsync",
        commander.Exact("-av"),
        commander.Regex(`.*\.tar\.gz$`),
        commander.Any())

    // Test your code here...
}
```

### Process Mocking
```go
func TestProcessMocking(t *testing.T) {
    mock := commander.NewMock()

    // Mock a streaming process
    mock.Expect("tail", "-f", "/var/log/messages").
        ReturnOutput([]byte("log line 1\nlog line 2\nlog line 3"))

    proc, err := mock.Start(context.Background(), "tail", []string{"-f", "/var/log/messages"})
    require.NoError(t, err)

    // Test streaming
    stdout := make(chan string, 10)
    proc.Stream(stdout, nil)

    var lines []string
    for line := range stdout {
        lines = append(lines, line)
    }

    expected := []string{"log line 1", "log line 2", "log line 3"}
    assert.Equal(t, expected, lines)

    require.NoError(t, mock.VerifyExpectations())
}
```

### Mock Utilities
```go
func TestMockUtilities(t *testing.T) {
    mock := commander.NewMock()

    // Set up multiple expectations
    mock.Expect("first").ReturnOutput([]byte("1"))
    mock.Expect("second").ReturnOutput([]byte("2"))
    mock.Expect("third").ReturnError(errors.New("failed"))

    // Execute them
    mock.Output(context.Background(), "first", nil)
    mock.Output(context.Background(), "second", nil)
    mock.Output(context.Background(), "third", nil)

    // Check call order
    order := mock.CallOrder()
    expected := []string{"first ", "second ", "third "}
    assert.Equal(t, expected, order)

    // Reset if needed
    mock.Reset() // Clears all expectations and history

    require.NoError(t, mock.VerifyExpectations())
}
```

## Performance and Concurrency

### Thread Safety
Everything is thread-safe. You can:
- Use the same Commander instance from multiple goroutines
- Run multiple commands concurrently 
- Stream from multiple processes simultaneously
- Use mocks in parallel tests

```go
func TestConcurrentMocking(t *testing.T) {
    mock := commander.NewMock()

    // Set up expectations for concurrent calls
    for i := 0; i < 10; i++ {
        mock.Expect("echo", string(rune('a'+i))).
            ReturnOutput([]byte(string(rune('A'+i))))
    }

    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(index int) {
            defer wg.Done()
            
            stdout, _, err := mock.Output(context.Background(), 
                "echo", []string{string(rune('a'+index))})
            
            require.NoError(t, err)
            assert.Equal(t, string(rune('A'+index)), string(stdout))
        }(i)
    }
    
    wg.Wait()
    require.NoError(t, mock.VerifyExpectations())
}
```

### Memory Management
- Channels are automatically closed when processes end
- Context cancellation is properly handled
- No memory leaks from goroutines or file descriptors
- Process cleanup uses `sync.Once` for safety

## Error Types Reference

Here's all the shit that can go wrong and how to handle it:

```go
// Package-specific errors
var (
    ErrUnexpectedCommand        = errors.New("unexpected command")
    ErrExpectedCommandNotCalled = errors.New("expected command not called")
    ErrProcessStartFailed       = errors.New("process start failed")
    ErrProcessWaitFailed        = errors.New("process wait failed")
    ErrPipeCreationFailed       = errors.New("pipe creation failed")
    ErrCommandFailed            = errors.New("command failed")
)

// Common errors (from github.com/psyb0t/common-go/errors)
commonerrors.ErrTimeout     // Command timed out
commonerrors.ErrTerminated  // Process terminated by SIGTERM
commonerrors.ErrKilled      // Process killed by SIGKILL
```

## Real-world Examples

### Deploy Script
```go
func deployApp(cmd commander.Commander) error {
    ctx := context.Background()

    fmt.Println("🏗️  Building application...")
    err := cmd.Run(ctx, "go", []string{"build", "-o", "app", "./cmd/server"})
    if err != nil {
        return fmt.Errorf("build failed: %w", err)
    }

    fmt.Println("🧪 Running tests...")
    testCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()
    err = cmd.Run(testCtx, "go", []string{"test", "./..."})
    if err != nil {
        return fmt.Errorf("tests failed: %w", err)
    }

    fmt.Println("📦 Creating Docker image...")
    err = cmd.Run(ctx, "docker", []string{"build", "-t", "myapp:latest", "."})
    if err != nil {
        return fmt.Errorf("docker build failed: %w", err)
    }

    fmt.Println("🚀 Pushing to registry...")
    pushCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
    defer cancel()
    err = cmd.Run(pushCtx, "docker", []string{"push", "myapp:latest"})
    if err != nil {
        return fmt.Errorf("docker push failed: %w", err)
    }

    fmt.Println("✅ Deploy completed successfully!")
    return nil
}
```

### Log Monitor
```go
func monitorLogs(cmd commander.Commander) error {
    ctx := context.Background()

    proc, err := cmd.Start(ctx, "tail", []string{"-f", "/var/log/app.log"})
    if err != nil {
        return fmt.Errorf("failed to start log monitoring: %w", err)
    }

    stdout := make(chan string, 100)
    proc.Stream(stdout, nil)

    // Monitor for specific patterns
    errorPattern := regexp.MustCompile(`(?i)error|exception|panic`)
    warningPattern := regexp.MustCompile(`(?i)warning|warn`)

    go func() {
        for line := range stdout {
            switch {
            case errorPattern.MatchString(line):
                log.Printf("🚨 ERROR: %s", line)
                // Maybe send alert, page someone, etc.
            case warningPattern.MatchString(line):
                log.Printf("⚠️  WARNING: %s", line)
            default:
                log.Printf("ℹ️  INFO: %s", line)
            }
        }
    }()

    // Stop monitoring after 1 hour
    time.Sleep(1 * time.Hour)
    
    stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    return proc.Stop(stopCtx)
}
```

### System Health Check
```go
func healthCheck(cmd commander.Commander) (map[string]bool, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    checks := map[string][]string{
        "disk_space":    {"df", "-h", "/"},
        "memory":        {"free", "-m"},
        "load_average":  {"uptime"},
        "docker":        {"docker", "ps"},
        "nginx":         {"systemctl", "is-active", "nginx"},
        "database":      {"pg_isready"},
    }

    results := make(map[string]bool)
    var wg sync.WaitGroup

    for name, cmdArgs := range checks {
        wg.Add(1)
        go func(checkName string, args []string) {
            defer wg.Done()
            
            err := cmd.Run(ctx, args[0], args[1:])
            results[checkName] = (err == nil)
            
            if err != nil {
                log.Printf("❌ Health check '%s' failed: %v", checkName, err)
            } else {
                log.Printf("✅ Health check '%s' passed", checkName)
            }
        }(name, cmdArgs)
    }

    wg.Wait()
    return results, nil
}
```

## Dependencies

- `github.com/sirupsen/logrus` - For debug logging
- `github.com/psyb0t/ctxerrors` - Error wrapping with context
- `github.com/psyb0t/common-go` - Common error types
- Standard library: `context`, `os/exec`, `sync`, `syscall`, etc.

## Why Use This?

### Before (stdlib `os/exec`)
```go
// Painful, error-prone, lots of boilerplate
cmd := exec.CommandContext(ctx, "some-command", "arg1", "arg2")
stdout, err := cmd.StdoutPipe()
if err != nil {
    // handle error
}
stderr, err := cmd.StderrPipe()
if err != nil {
    // handle error
}

err = cmd.Start()
if err != nil {
    // handle error
}

// Now you need to read from pipes in goroutines...
// And handle timeouts manually...
// And figure out why your process is hanging...
// And write your own mocks...
// 🤮
```

### After (Commander)
```go
// Clean, simple, powerful
cmd := commander.New()
stdout, stderr, err := cmd.Output(ctx, "some-command", []string{"arg1", "arg2"})
if err != nil {
    // handle error (with proper context!)
}
// Done. That's it. 🎉
```

## License

MIT - Use it, abuse it, whatever. Just don't blame anyone if your servers catch fire. 🔥

## Contributing

Found a bug? Want a feature? Open an issue or send a PR. Contributing is welcome.

---

**Commander: SOMEBODY STOP ME from this beautiful command-line carnage! P-A-R-T-Y time for your digital violence! 🔥👹💚⚰️** 🐚