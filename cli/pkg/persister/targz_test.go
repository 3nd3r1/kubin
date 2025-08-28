package persister

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/3nd3r1/kubin/cli/pkg/collector"
)

func TestTarGzPersister_GenerateSnapshotFilename(t *testing.T) {
	persister, err := NewTarGzPersister()
	if err != nil {
		t.Fatalf("Failed to create persister: %v", err)
	}
	defer persister.cleanup()

	// Test filename format
	filename := persister.generateOutputFilename()

	// Check that filename matches expected pattern: kubin-snapshot-{timestamp}-{nanoseconds}.tar.gz
	expectedPattern := `^kubin-snapshot-\d+-\d{9}\.tar\.gz$`
	matched, err := regexp.MatchString(expectedPattern, filename)
	if err != nil {
		t.Fatalf("Failed to compile regex: %v", err)
	}
	if !matched {
		t.Errorf("Filename %s does not match expected pattern %s", filename, expectedPattern)
	}

	t.Logf("Generated filename: %s", filename)
}

func TestTarGzPersister_FilenameUniqueness(t *testing.T) {
	persister, err := NewTarGzPersister()
	if err != nil {
		t.Fatalf("Failed to create persister: %v", err)
	}
	defer persister.cleanup()

	filenames := make(map[string]bool)
	for range 100 {
		filename := persister.generateOutputFilename()

		// Check for collisions
		if filenames[filename] {
			t.Errorf("Collision detected: filename %s was generated twice", filename)
		}
		filenames[filename] = true

		// No sleep needed since nanoseconds provide sufficient precision
	}

	if len(filenames) != 100 {
		t.Errorf("Expected 100 unique filenames, got %d", len(filenames))
	}

	t.Logf("Generated %d unique filenames", len(filenames))
}

func TestTarGzPersister_TimestampFormat(t *testing.T) {
	persister, err := NewTarGzPersister()
	if err != nil {
		t.Fatalf("Failed to create persister: %v", err)
	}
	defer persister.cleanup()

	// Record time before generating filename
	beforeTime := time.Now().UTC()
	filename := persister.generateOutputFilename()
	afterTime := time.Now().UTC()

	// Extract timestamp and nanoseconds from filename
	var extractedTimestamp int64
	var extractedNanoseconds int
	n, err := fmt.Sscanf(filename, "kubin-snapshot-%d-%d.tar.gz", &extractedTimestamp, &extractedNanoseconds)
	if err != nil || n != 2 {
		t.Fatalf("Failed to extract timestamp and nanoseconds from filename %s: %v", filename, err)
	}

	// Verify timestamp is within reasonable range
	if extractedTimestamp < beforeTime.Unix() || extractedTimestamp > afterTime.Unix() {
		t.Errorf("Timestamp %d is not between %d and %d", extractedTimestamp, beforeTime.Unix(), afterTime.Unix())
	}

	// Verify nanoseconds are valid (0-999999999)
	if extractedNanoseconds < 0 || extractedNanoseconds > 999999999 {
		t.Errorf("Nanoseconds %d is not valid (should be 0-999999999)", extractedNanoseconds)
	}

	// Verify timestamp represents a reasonable date (after 2020)
	minValidTimestamp := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	if extractedTimestamp < minValidTimestamp {
		t.Errorf("Timestamp %d appears to be invalid (before 2020)", extractedTimestamp)
	}

	t.Logf("Extracted timestamp: %d, nanoseconds: %d (represents %s)",
		extractedTimestamp, extractedNanoseconds,
		time.Unix(extractedTimestamp, int64(extractedNanoseconds)).UTC().Format(time.RFC3339Nano))
}

func TestTarGzPersister_FinalizeCreatesFile(t *testing.T) {
	persister, err := NewTarGzPersister()
	if err != nil {
		t.Fatalf("Failed to create persister: %v", err)
	}

	// Add some test data
	testResource := collector.ClusterResource{
		Kind: "Pod",
		Name: "test-pod",
		Data: map[string]any{
			"metadata": map[string]any{
				"name":      "test-pod",
				"namespace": "default",
			},
		},
		Metadata: map[string]string{
			"namespace": "default",
		},
	}

	err = persister.Persist(testResource)
	if err != nil {
		t.Fatalf("Failed to persist resource: %v", err)
	}

	// Record files before finalize
	filesBefore, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	// Finalize
	err = persister.Finalize()
	if err != nil {
		t.Fatalf("Failed to finalize: %v", err)
	}

	// Check that a new tar.gz file was created
	filesAfter, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	// Find the new file
	var newFiles []string
	filesBeforeMap := make(map[string]bool)
	for _, file := range filesBefore {
		filesBeforeMap[file.Name()] = true
	}

	expectedPattern := `^kubin-snapshot-\d+-\d{9}\.tar\.gz$`
	var snapshotFile string

	for _, file := range filesAfter {
		if !filesBeforeMap[file.Name()] {
			newFiles = append(newFiles, file.Name())
			match, err := regexp.Compile(expectedPattern)
			if err != nil {
				t.Fatalf("Failed to compile regex: %v", err)
			}
			if match.MatchString(file.Name()) {
				snapshotFile = file.Name()
			}
		}
	}

	if snapshotFile == "" {
		t.Errorf("No snapshot file matching pattern %s was created. New files: %v", expectedPattern, newFiles)
	} else {
		// Verify the file exists and has content
		fileInfo, err := os.Stat(snapshotFile)
		if err != nil {
			t.Errorf("Snapshot file does not exist: %v", err)
		} else if fileInfo.Size() == 0 {
			t.Errorf("Snapshot file is empty")
		}

		// Clean up
		os.Remove(snapshotFile)
		t.Logf("Successfully created snapshot: %s (size: %d bytes)", snapshotFile, fileInfo.Size())
	}
}

func TestTarGzPersister_TimezoneConsistency(t *testing.T) {
	persister, err := NewTarGzPersister()
	if err != nil {
		t.Fatalf("Failed to create persister: %v", err)
	}
	defer persister.cleanup()

	// Test that filename generation is consistent regardless of local timezone
	originalTZ := time.Local
	defer func() { time.Local = originalTZ }()

	// Test with different timezones
	timezones := []string{"UTC", "America/New_York", "Europe/London", "Asia/Tokyo"}

	for _, tzName := range timezones {
		if tzName == "UTC" {
			time.Local = time.UTC
		} else {
			tz, err := time.LoadLocation(tzName)
			if err != nil {
				t.Skipf("Could not load timezone %s: %v", tzName, err)
				continue
			}
			time.Local = tz
		}

		filename := persister.generateOutputFilename()
		t.Logf("Timezone %s: %s", tzName, filename)

		match, err := regexp.Compile(`^kubin-snapshot-\d+-\d{9}\.tar\.gz$`)
		if err != nil {
			t.Fatalf("Failed to compile regex: %v", err)
		}

		if !match.MatchString(filename) {
			t.Errorf("Filename %s does not match expected pattern", filename)
		}
	}
}
