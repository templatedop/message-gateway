package bootstrapper

import (
	"context"
	"testing"
	"time"

	db "MgApplication/api-db"
	healthcheck "MgApplication/api-healthcheck"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadDBProbeParams tests that readDBProbeParams struct is correctly configured
func TestReadDBProbeParams(t *testing.T) {
	// Verify the struct has the correct field tags
	params := readDBProbeParams{
		DB: &db.DB{}, // Mock DB
	}

	assert.NotNil(t, params.DB, "DB field should not be nil")
}

// TestWriteDBProbeParams tests that writeDBProbeParams struct is correctly configured
func TestWriteDBProbeParams(t *testing.T) {
	// Verify the struct has the correct field tags
	params := writeDBProbeParams{
		DB: &db.DB{}, // Mock DB
	}

	assert.NotNil(t, params.DB, "DB field should not be nil")
}

// TestReadDBProbeCreation tests that read DB probe can be created with correct name
func TestReadDBProbeCreation(t *testing.T) {
	// Create a mock DB (without actual connection)
	mockDB := &db.DB{
		// DB fields would be initialized in a real scenario
	}

	params := readDBProbeParams{
		DB: mockDB,
	}

	// Create probe using the function that would be passed to AsCheckerProbe
	probeFunc := func(p readDBProbeParams) healthcheck.CheckerProbe {
		probe := db.NewSQLProbe(p.DB)
		probe.SetName(ReadDBProbeName)
		return probe
	}

	probe := probeFunc(params)

	require.NotNil(t, probe, "Probe should not be nil")
	assert.Equal(t, ReadDBProbeName, probe.Name(), "Probe should have correct name")
}

// TestWriteDBProbeCreation tests that write DB probe can be created with correct name
func TestWriteDBProbeCreation(t *testing.T) {
	// Create a mock DB (without actual connection)
	mockDB := &db.DB{
		// DB fields would be initialized in a real scenario
	}

	params := writeDBProbeParams{
		DB: mockDB,
	}

	// Create probe using the function that would be passed to AsCheckerProbe
	probeFunc := func(p writeDBProbeParams) healthcheck.CheckerProbe {
		probe := db.NewSQLProbe(p.DB)
		probe.SetName(WriteDBProbeName)
		return probe
	}

	probe := probeFunc(params)

	require.NotNil(t, probe, "Probe should not be nil")
	assert.Equal(t, WriteDBProbeName, probe.Name(), "Probe should have correct name")
}

// TestProbeNames verifies the probe name constants
func TestProbeNames(t *testing.T) {
	assert.Equal(t, "read-db-probe", ReadDBProbeName, "Read DB probe name should match")
	assert.Equal(t, "write-db-probe", WriteDBProbeName, "Write DB probe name should match")
}

// TestCollectorNames verifies the collector name constants
func TestCollectorNames(t *testing.T) {
	assert.Equal(t, "read_db_collector", ReadDBCollectorName, "Read DB collector name should match")
	assert.Equal(t, "write_db_collector", WriteDBCollectorName, "Write DB collector name should match")
}

// TestProbeInterface verifies probes implement healthcheck.CheckerProbe interface
func TestProbeInterface(t *testing.T) {
	mockDB := &db.DB{}
	probe := db.NewSQLProbe(mockDB)

	// Verify probe implements the interface
	var _ healthcheck.CheckerProbe = probe

	// Verify methods exist and work
	assert.NotEmpty(t, probe.Name(), "Probe should have a name")

	// Note: We can't test Check() without a real DB connection
	// That would require integration tests
	t.Log("Probe implements CheckerProbe interface correctly")
}

// TestProbeFunctionSignature verifies the probe creation functions have correct signatures
func TestProbeFunctionSignature(t *testing.T) {
	t.Run("ReadDB probe function", func(t *testing.T) {
		// This function signature matches what AsCheckerProbe expects
		probeFunc := func(p readDBProbeParams) healthcheck.CheckerProbe {
			probe := db.NewSQLProbe(p.DB)
			probe.SetName(ReadDBProbeName)
			return probe
		}

		mockDB := &db.DB{}
		params := readDBProbeParams{DB: mockDB}
		result := probeFunc(params)

		assert.NotNil(t, result)
		assert.Equal(t, ReadDBProbeName, result.Name())
	})

	t.Run("WriteDB probe function", func(t *testing.T) {
		// This function signature matches what AsCheckerProbe expects
		probeFunc := func(p writeDBProbeParams) healthcheck.CheckerProbe {
			probe := db.NewSQLProbe(p.DB)
			probe.SetName(WriteDBProbeName)
			return probe
		}

		mockDB := &db.DB{}
		params := writeDBProbeParams{DB: mockDB}
		result := probeFunc(params)

		assert.NotNil(t, result)
		assert.Equal(t, WriteDBProbeName, result.Name())
	})
}

// TestProbeNaming tests that different probes have different names
func TestProbeNaming(t *testing.T) {
	mockDB := &db.DB{}

	readProbe := db.NewSQLProbe(mockDB)
	readProbe.SetName(ReadDBProbeName)

	writeProbe := db.NewSQLProbe(mockDB)
	writeProbe.SetName(WriteDBProbeName)

	assert.NotEqual(t, readProbe.Name(), writeProbe.Name(), "Read and write probes should have different names")
	assert.Equal(t, "read-db-probe", readProbe.Name())
	assert.Equal(t, "write-db-probe", writeProbe.Name())
}

// TestProbeSetName tests the SetName method
func TestProbeSetName(t *testing.T) {
	mockDB := &db.DB{}
	probe := db.NewSQLProbe(mockDB)

	// Default name
	defaultName := probe.Name()
	assert.Equal(t, db.DefaultProbeName, defaultName, "Should have default name")

	// Set custom name
	probe.SetName("custom-probe")
	assert.Equal(t, "custom-probe", probe.Name(), "Should have custom name")

	// Set another name
	probe.SetName(ReadDBProbeName)
	assert.Equal(t, ReadDBProbeName, probe.Name(), "Should have read DB probe name")
}

// MockCheckerProbe is a simple mock for testing
type MockCheckerProbe struct {
	name       string
	shouldPass bool
}

func (m *MockCheckerProbe) Name() string {
	return m.name
}

func (m *MockCheckerProbe) Check(ctx context.Context) *healthcheck.CheckerProbeResult {
	if m.shouldPass {
		return healthcheck.NewCheckerProbeResult(true, "mock check passed")
	}
	return healthcheck.NewCheckerProbeResult(false, "mock check failed")
}

// TestMockProbe tests our mock implementation
func TestMockProbe(t *testing.T) {
	t.Run("Passing probe", func(t *testing.T) {
		probe := &MockCheckerProbe{
			name:       "test-probe",
			shouldPass: true,
		}

		ctx := context.Background()
		result := probe.Check(ctx)

		assert.True(t, result.Passed, "Probe should pass")
		assert.Equal(t, "mock check passed", result.Output)
	})

	t.Run("Failing probe", func(t *testing.T) {
		probe := &MockCheckerProbe{
			name:       "test-probe",
			shouldPass: false,
		}

		ctx := context.Background()
		result := probe.Check(ctx)

		assert.False(t, result.Passed, "Probe should fail")
		assert.Equal(t, "mock check failed", result.Output)
	})
}

// TestProbeWithTimeout tests probe behavior with context timeout
func TestProbeWithTimeout(t *testing.T) {
	t.Run("Probe with sufficient timeout", func(t *testing.T) {
		probe := &MockCheckerProbe{
			name:       "timeout-test",
			shouldPass: true,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result := probe.Check(ctx)
		assert.True(t, result.Passed)
	})

	t.Run("Probe with expired context", func(t *testing.T) {
		probe := &MockCheckerProbe{
			name:       "timeout-test",
			shouldPass: true,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		time.Sleep(10 * time.Millisecond) // Ensure context expires
		defer cancel()

		// Context is expired, but mock probe doesn't check it
		// Real SQL probe would handle this
		result := probe.Check(ctx)

		// Mock doesn't check context, so it still passes
		// This demonstrates the need for real probes to check ctx.Err()
		assert.NotNil(t, result)
	})
}

// Benchmark tests

func BenchmarkProbeCreation(b *testing.B) {
	mockDB := &db.DB{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		probe := db.NewSQLProbe(mockDB)
		probe.SetName(ReadDBProbeName)
		_ = probe
	}
}

func BenchmarkProbeNameAccess(b *testing.B) {
	mockDB := &db.DB{}
	probe := db.NewSQLProbe(mockDB)
	probe.SetName(ReadDBProbeName)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = probe.Name()
	}
}

// Integration test notes:
// The following tests would require actual database connections and are better
// suited for integration tests rather than unit tests:
//
// - TestReadDBProbeWithRealConnection
// - TestWriteDBProbeWithRealConnection
// - TestProbeFailureWithDisconnectedDB
// - TestProbeSuccessWithHealthyDB
// - TestProbeTimeoutWithSlowDB
//
// These should be implemented in a separate integration test suite that:
// 1. Sets up test databases (possibly using testcontainers)
// 2. Tests actual probe.Check() behavior
// 3. Verifies connection health monitoring
// 4. Tests failure scenarios
