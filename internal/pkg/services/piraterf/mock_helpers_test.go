package piraterf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/psyb0t/commander"
	"github.com/psyb0t/ctxerrors"
)

// Simple mock helpers for compatibility across test files.
func newMockCommander() *commander.MockCommander {
	return commander.NewMock()
}

func newMockProcess() commander.Process {
	mock := commander.NewMock()
	proc, _ := mock.Start(context.Background(), "test", []string{})

	return proc
}

// fileCreatingMockCommander is a wrapper around MockCommander that creates
// output files for ffmpeg and convert commands to simulate successful conversion.
type fileCreatingMockCommander struct {
	commander.MockCommander
}

func (m *fileCreatingMockCommander) Start(
	ctx context.Context,
	name string,
	args []string,
	opts ...commander.Option,
) (commander.Process, error) {
	// Call the original Start method
	proc, err := m.MockCommander.Start(ctx, name, args, opts...)
	if err != nil {
		return proc, err
	}

	// If this is an ffmpeg command, create the output file
	if name == "ffmpeg" && len(args) >= 10 {
		outputPath := args[len(args)-1] // Last argument is output path
		// Create the directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return proc, err
		}
		// Create the output file to simulate ffmpeg success
		if err := os.WriteFile(outputPath, []byte("fake wav content"), 0o644); err != nil {
			return proc, err
		}
	}

	// If this is a convert (ImageMagick) command, create the output files
	if name == "convert" && len(args) >= 11 {
		outputPath := args[len(args)-1] // Last argument is output path (base.yuv)
		// Create the directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return proc, err
		}

		// For partition interlace, ImageMagick creates separate .Y, .U, .V files
		// We need to create the .Y file that the code expects
		basePath := outputPath[:len(outputPath)-4] // Remove .yuv extension

		yFilePath := basePath + ".Y"
		if err := os.WriteFile(yFilePath, []byte("fake Y channel content"), 0o644); err != nil {
			return proc, err
		}

		// Optionally create .U and .V files too for completeness
		uFilePath := basePath + ".U"
		if err := os.WriteFile(uFilePath, []byte("fake U channel content"), 0o644); err != nil {
			return proc, err
		}

		vFilePath := basePath + ".V"
		if err := os.WriteFile(vFilePath, []byte("fake V channel content"), 0o644); err != nil {
			return proc, err
		}
	}

	// If this is a sox command, create output files when needed
	if name == "sox" && len(args) >= 3 {
		// For sox silence addition (audioFile, outputFile, "pad", "0", "2")
		if len(args) == 5 && args[2] == "pad" && args[3] == "0" && args[4] == "2" {
			outputPath := args[1] // Second argument is output path
			// Create the directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
				return proc, err
			}
			// Create the output file to simulate sox success
			if err := os.WriteFile(outputPath, []byte("fake audio with silence"), 0o644); err != nil {
				return proc, err
			}
		}
		// For sox playlist creation (multiple input files + output file)
		if len(args) >= 4 {
			// Check if this looks like a playlist creation (last arg is output, others are inputs or options)
			lastArg := args[len(args)-1]
			if filepath.Ext(lastArg) == ".wav" || filepath.Ext(lastArg) == ".mp3" {
				// Create the directory if it doesn't exist
				if err := os.MkdirAll(filepath.Dir(lastArg), 0o755); err != nil {
					return proc, err
				}
				// Create the output file to simulate sox success
				if err := os.WriteFile(lastArg, []byte("fake playlist audio"), 0o644); err != nil {
					return proc, err
				}
			}
		}
	}

	return proc, err
}

// testMockCommander handles sox and convert commands directly for testing.
type testMockCommander struct {
	tempDir string
}

func (m *testMockCommander) Start(
	ctx context.Context,
	name string,
	args []string,
	opts ...commander.Option,
) (commander.Process, error) {
	// Handle sox silence addition: sox audioFile outputFile pad 0 2
	if name == "sox" && len(args) == 5 && args[2] == "pad" && args[3] == "0" && args[4] == "2" {
		outputPath := args[1] // Second argument is output path
		// Create the directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return nil, err
		}
		// Create the output file to simulate sox success
		if err := os.WriteFile(outputPath, []byte("fake audio with silence"), 0o644); err != nil {
			return nil, err
		}
		// Return a successful mock process
		mock := commander.NewMock()
		mock.Expect("dummy").ReturnOutput([]byte(""))
		proc, _ := mock.Start(context.Background(), "dummy", []string{})

		return proc, nil
	}

	// Handle sox playlist creation: sox file1 file2 file3 -r rate -b bits -c channels outputFile
	if name == "sox" && len(args) >= 8 {
		// Find the output file (last argument)
		outputPath := args[len(args)-1]
		// Create the directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return nil, err
		}
		// Create the output file to simulate sox success
		if err := os.WriteFile(outputPath, []byte("fake playlist audio"), 0o644); err != nil {
			return nil, err
		}
		// Return a successful mock process
		mock := commander.NewMock()
		mock.Expect("dummy").ReturnOutput([]byte(""))
		proc, _ := mock.Start(context.Background(), "dummy", []string{})

		return proc, nil
	}

	// Handle convert (ImageMagick) command for image conversion
	if name == "convert" && len(args) >= 11 {
		outputPath := args[len(args)-1] // Last argument is output path (base.yuv)
		// Create the directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return nil, err
		}

		// For partition interlace, ImageMagick creates separate .Y, .U, .V files
		// We need to create the .Y file that the code expects
		basePath := outputPath[:len(outputPath)-4] // Remove .yuv extension

		yFilePath := basePath + ".Y"
		if err := os.WriteFile(yFilePath, []byte("fake Y channel content"), 0o644); err != nil {
			return nil, err
		}

		// Optionally create .U and .V files too for completeness
		uFilePath := basePath + ".U"
		if err := os.WriteFile(uFilePath, []byte("fake U channel content"), 0o644); err != nil {
			return nil, err
		}

		vFilePath := basePath + ".V"
		if err := os.WriteFile(vFilePath, []byte("fake V channel content"), 0o644); err != nil {
			return nil, err
		}

		// Return a successful mock process
		mock := commander.NewMock()
		mock.Expect("dummy").ReturnOutput([]byte(""))
		proc, _ := mock.Start(context.Background(), "dummy", []string{})

		return proc, nil
	}

	return nil, ctxerrors.New("unexpected command: " + name)
}

func (m *testMockCommander) Run(ctx context.Context, name string, args []string, opts ...commander.Option) error {
	_, err := m.Start(ctx, name, args, opts...)

	return err
}

func (m *testMockCommander) Output(ctx context.Context, name string, args []string, opts ...commander.Option) ([]byte, []byte, error) {
	// Handle sox duration check: sox --info -D audioFile
	if name == "sox" && len(args) == 3 && args[0] == "--info" && args[1] == "-D" {
		if args[2] == ".fixtures/test_3s.wav" {
			return []byte("3.000000"), []byte(""), nil
		}

		if args[2] == "test.wav" {
			return []byte("3.500000"), []byte(""), nil
		}
		// Default for any audio file
		return []byte("3.000000"), []byte(""), nil
	}

	return nil, nil, ctxerrors.New("unexpected command: " + name + " with args: " + fmt.Sprintf("%v", args))
}

func (m *testMockCommander) CombinedOutput(ctx context.Context, name string, args []string, opts ...commander.Option) ([]byte, error) {
	stdout, _, err := m.Output(ctx, name, args, opts...)

	return stdout, err
}
