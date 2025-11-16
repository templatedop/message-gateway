package db

import (
	"context"
	"fmt"
	"strings"

	healthcheck "MgApplication/api-healthcheck"

	log "MgApplication/api-log"
)

const DefaultProbeName = "Database"

type SQLProbe struct {
	name string
	db   *DB
}

// NewSQLProbe returns a new [SQLProbe].
func NewSQLProbe(db *DB) *SQLProbe {
	return &SQLProbe{
		name: DefaultProbeName,
		db:   db,
	}
}

// Name returns the name of the [SQLProbe].
func (p *SQLProbe) Name() string {
	return p.name
}

// SetName sets the name of the [SQLProbe].
func (p *SQLProbe) SetName(name string) *SQLProbe {
	p.name = name

	return p
}
func isSocketError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "socket") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "hung up")
}

// Check returns a successful [healthcheck.CheckerProbeResult] if the database connection can be pinged.
func (p *SQLProbe) Check(ctx context.Context) (result *healthcheck.CheckerProbeResult) {

	err := p.db.Pool.Ping(ctx)

	if err != nil {

		log.GetBaseLoggerInstance().ToZerolog().Error().Err(err).Msg("database ping error")

		switch {
		case err == context.DeadlineExceeded || ctx.Err() == context.DeadlineExceeded:
			log.GetBaseLoggerInstance().ToZerolog().Error().
				Str("probe", p.name).
				Err(err).
				Msg("database ping timeout")

			result = healthcheck.NewCheckerProbeResult(false, "database connection timeout after 10s")
			return
		case isSocketError(err):
			log.GetBaseLoggerInstance().ToZerolog().Error().
				Str("probe", p.name).
				Err(err).
				Msg("socket connection error")
			result = healthcheck.NewCheckerProbeResult(false, "database socket connection error")
			return
		default:
			log.GetBaseLoggerInstance().ToZerolog().Error().
				Str("probe", p.name).
				Err(err).
				Msg("database ping failed")
			result = healthcheck.NewCheckerProbeResult(false, fmt.Sprintf("database connection failed: %v", err))
			return
		}
		//return healthcheck.NewCheckerProbeResult(false, fmt.Sprintf("database ping error: %v", err))
	}

	return healthcheck.NewCheckerProbeResult(true, "database ping success")
}
