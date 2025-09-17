package commander

import (
	"bytes"
	"context"
	"io"
	"regexp"
	"strings"
	"sync"

	"github.com/psyb0t/ctxerrors"
)

// ArgumentMatcher interface for flexible argument matching
type ArgumentMatcher interface {
	Matches(arg string) bool
	String() string
}

// Exact matcher
type ExactMatcher struct {
	expected string
}

func (m *ExactMatcher) Matches(arg string) bool {
	return m.expected == arg
}

func (m *ExactMatcher) String() string {
	return m.expected
}

// Regex matcher
type RegexMatcher struct {
	pattern *regexp.Regexp
	raw     string
}

func (m *RegexMatcher) Matches(arg string) bool {
	return m.pattern.MatchString(arg)
}

func (m *RegexMatcher) String() string {
	return "regex:" + m.raw
}

// Any matcher
type AnyMatcher struct{}

func (m *AnyMatcher) Matches(_ string) bool {
	return true
}

func (m *AnyMatcher) String() string {
	return "*"
}

// Helper functions for creating matchers
func Exact(s string) ArgumentMatcher { //nolint:ireturn
	// factory function for matcher interface
	return &ExactMatcher{expected: s}
}

func Regex(pattern string) ArgumentMatcher { //nolint:ireturn
	// factory function for matcher interface
	return &RegexMatcher{
		pattern: regexp.MustCompile(pattern),
		raw:     pattern,
	}
}

func Any() ArgumentMatcher { //nolint:ireturn
	// factory function for matcher interface
	return &AnyMatcher{}
}

// MockCommander for testing
type MockCommander struct {
	expectations []Expectation
	mu           sync.Mutex
	callOrder    []string // track call order
}

type Expectation struct {
	Name     string
	Args     []string
	Matchers []ArgumentMatcher
	Output   []byte
	Error    error
	Called   bool
}

func NewMock() *MockCommander {
	return &MockCommander{
		expectations: make([]Expectation, 0),
		callOrder:    make([]string, 0),
	}
}

func (m *MockCommander) Expect(
	name string,
	args ...string,
) *Expectation {
	m.mu.Lock()
	defer m.mu.Unlock()

	exp := Expectation{
		Name: name,
		Args: args,
	}
	m.expectations = append(m.expectations, exp)

	return &m.expectations[len(m.expectations)-1]
}

func (m *MockCommander) ExpectWithMatchers(
	name string,
	matchers ...ArgumentMatcher,
) *Expectation {
	m.mu.Lock()
	defer m.mu.Unlock()

	exp := Expectation{
		Name:     name,
		Matchers: matchers,
	}
	m.expectations = append(m.expectations, exp)

	return &m.expectations[len(m.expectations)-1]
}

func (m *MockCommander) Run(
	_ context.Context,
	name string,
	args []string,
	_ ...Option,
) error {
	_, err := m.execute(name, args)

	return err
}

func (m *MockCommander) Output(
	_ context.Context,
	name string,
	args []string,
	_ ...Option,
) ([]byte, []byte, error) {
	output, execErr := m.execute(name, args)
	if execErr != nil {
		return nil, nil, execErr
	}

	// For mocking, we'll treat the output as both stdout and stderr
	return output, output, nil
}

func (m *MockCommander) CombinedOutput(
	_ context.Context,
	name string,
	args []string,
	_ ...Option,
) ([]byte, error) {
	output, execErr := m.execute(name, args)
	if execErr != nil {
		return nil, execErr
	}

	return output, nil
}

//nolint:ireturn // interface return by design
func (m *MockCommander) Start(
	_ context.Context,
	name string,
	args []string,
	_ ...Option,
) (Process, error) {
	output, err := m.execute(name, args)
	if err != nil {
		return nil, err
	}

	mockProc := &mockProcess{output: output}
	// Convert output to stream lines for streaming functionality
	if len(output) > 0 {
		lines := strings.Split(string(output), "\n")
		// Remove empty last line if present
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}

		mockProc.SetStreamOutput(lines)
	}

	return mockProc, nil
}

func (m *MockCommander) execute(
	name string,
	args []string,
) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callOrder = append(m.callOrder, name+" "+argsToString(args))

	for i := range m.expectations {
		exp := &m.expectations[i]
		if exp.Called {
			continue
		}

		if exp.matches(name, args) {
			exp.Called = true

			return exp.Output, exp.Error
		}
	}

	return nil, ctxerrors.Wrap(
		ErrUnexpectedCommand,
		"unexpected command",
	)
}

func (m *MockCommander) VerifyExpectations() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, exp := range m.expectations {
		if !exp.Called {
			return ctxerrors.Wrap(
				ErrExpectedCommandNotCalled,
				"expected command not called",
			)
		}
	}

	return nil
}

func (m *MockCommander) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.expectations = make([]Expectation, 0)
	m.callOrder = make([]string, 0)
}

func (m *MockCommander) CallOrder() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]string, len(m.callOrder))
	copy(result, m.callOrder)

	return result
}

// Expectation methods
func (e *Expectation) ReturnOutput(output []byte) *Expectation {
	e.Output = output

	return e
}

func (e *Expectation) ReturnError(err error) *Expectation {
	e.Error = err

	return e
}

func (e *Expectation) matches(
	name string,
	args []string,
) bool {
	if e.Name != name {
		return false
	}

	// Use matchers if available
	if len(e.Matchers) > 0 {
		if len(e.Matchers) != len(args) {
			return false
		}

		for i, matcher := range e.Matchers {
			if !matcher.Matches(args[i]) {
				return false
			}
		}

		return true
	}

	// Exact matching
	if len(e.Args) != len(args) {
		return false
	}

	for i := range e.Args {
		if e.Args[i] != args[i] {
			return false
		}
	}

	return true
}

// mockProcess for testing
type mockProcess struct {
	output      []byte
	streamLines []string
	streamIndex int
	streamMu    sync.Mutex
	stopped     bool
}

func (p *mockProcess) Start() error {
	return nil
}

func (p *mockProcess) Wait() error {
	return nil
}

func (p *mockProcess) StdinPipe() (io.WriteCloser, error) {
	return &nopWriteCloser{&bytes.Buffer{}}, nil
}

type nopWriteCloser struct {
	*bytes.Buffer
}

func (n *nopWriteCloser) Close() error {
	return nil
}

func (p *mockProcess) Stream(
	stdout, stderr chan<- string,
) {
	go func() {
		if stdout != nil {
			defer close(stdout)
		}

		if stderr != nil {
			defer close(stderr)
		}

		p.streamMu.Lock()
		startIndex := p.streamIndex
		lines := make([]string, len(p.streamLines)-startIndex)
		copy(lines, p.streamLines[startIndex:])
		p.streamMu.Unlock()

		// Send all available lines to stdout channel
		// (mock assumes stdout)
		for _, line := range lines {
			if stdout != nil {
				select {
				case stdout <- line:
					// Line sent successfully
				default:
					// Channel is closed or blocked
					return
				}
			}
		}
	}()
}

// SetStreamOutput configures the mock process to stream specific lines
func (p *mockProcess) SetStreamOutput(lines []string) {
	p.streamMu.Lock()
	defer p.streamMu.Unlock()

	p.streamLines = lines
	p.streamIndex = 0
}

// SimulateStreamLine adds a new line to the stream (simulates live output)
func (p *mockProcess) SimulateStreamLine(line string) {
	p.streamMu.Lock()
	defer p.streamMu.Unlock()

	p.streamLines = append(p.streamLines, line)
}

func (p *mockProcess) Stop(_ context.Context) error {
	p.streamMu.Lock()
	defer p.streamMu.Unlock()

	p.stopped = true

	return nil
}

func (p *mockProcess) Kill(ctx context.Context) error {
	return p.Stop(ctx)
}

const mockPID = 99999

func (p *mockProcess) PID() int {
	// Mock processes don't have real PIDs
	return mockPID
}

// Utility functions
func argsToString(args []string) string {
	if len(args) == 0 {
		return ""
	}

	result := args[0]

	for i := 1; i < len(args); i++ {
		result += " " + args[i]
	}

	return result
}
