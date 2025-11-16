package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	//"treasurymanagement/healthcheck"
	healthcheck "MgApplication/api-healthcheck"

	log "MgApplication/api-log"

	"github.com/gin-gonic/gin"
)

func HealthCheckHandler(checker *healthcheck.Checker, kind healthcheck.ProbeKind) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 9*time.Second)
		defer cancel()
		result := checker.Check(ctx, kind)
		status := http.StatusOK

		if !result.Success {
			status = http.StatusServiceUnavailable

			evt := log.GetBaseLoggerInstance().ToZerolog().Error()

			for probeName, probeResult := range result.ProbesResults {
				evt.Str(probeName, fmt.Sprintf("success: %v, message: %s", probeResult.Success, probeResult.Message))
			}

			evt.Msg("healthcheck failure")

		}

		result.ProbesResults["Router"] = &healthcheck.CheckerProbeResult{
			Success: true,
			Message: "Router Health check Successful",
		}

		c.JSON(status, result)
	}

}

func MultipleHealthCheckHandler(checker *healthcheck.Checker, kind healthcheck.ProbeKind) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 9*time.Second)
		defer cancel()

		result := checker.Check(ctx, kind)
		status := http.StatusOK

		// Log all probe results for debugging
		evt := log.GetBaseLoggerInstance().ToZerolog().Info()
		if !result.Success {
			status = http.StatusServiceUnavailable
			evt = log.GetBaseLoggerInstance().ToZerolog().Error()
		}

		// Check for expected probes
		expectedProbes := []string{"write_db_probe", "read_db_probe"}
		for _, probeName := range expectedProbes {
			probeResult, exists := result.ProbesResults[probeName]
			if !exists {
				evt.Str(probeName, "probe not found")
				result.Success = false
				status = http.StatusServiceUnavailable
			} else {
				evt.Str(probeName, fmt.Sprintf("success: %v, message: %s", probeResult.Success, probeResult.Message))
			}
		}

		if !result.Success {
			evt.Msg("healthcheck failure")
		} else {
			evt.Msg("healthcheck success")
		}

		// Add Router probe result
		result.ProbesResults["Router"] = &healthcheck.CheckerProbeResult{
			Success: true,
			Message: "Router Health check Successful",
		}

		c.JSON(status, result)
	}
}
