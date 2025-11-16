package apierrors

// dbError represents a database-specific error code and message.
type dbError struct {
	Code    string
	Message string
}

var (
	// Success case (200 range)
	DBSuccessfulCompletion = dbError{Code: "00", Message: "Successful Completion"}

	// Client-side errors (400 range)
	DBNoData                                  = dbError{Code: "02", Message: "No Data"}
	DBWarning                                 = dbError{Code: "01", Message: "Warning"}                                     // HTTP 400 Bad Request: Something is wrong with the request, though not necessarily fatal.
	DBSQLStatementNotYetComplete              = dbError{Code: "03", Message: "SQL Statement Not Yet Complete"}              // HTTP 400 Bad Request: The SQL query is incomplete.
	DBInvalidGrantor                          = dbError{Code: "0L", Message: "Invalid Grantor"}                             // HTTP 403 Forbidden: Authorization issues or invalid privileges.
	DBInvalidRoleSpecification                = dbError{Code: "0P", Message: "Invalid Role Specification"}                  // HTTP 400 Bad Request: Role provided in the request is invalid.
	DBCaseNotFound                            = dbError{Code: "20", Message: "Case Not Found"}                              // HTTP 404 Not Found: Requested data or case could not be found.
	DBCardinalityViolation                    = dbError{Code: "21", Message: "Cardinality Violation"}                       // HTTP 400 Bad Request: Data does not meet constraints (e.g., multiple rows where one expected).
	DBDataException                           = dbError{Code: "22", Message: "Data Exception"}                              // HTTP 422 Unprocessable Entity: The request contains invalid data.
	DBIntegrityConstraintViolation            = dbError{Code: "23", Message: "Integrity Constraint Violation"}              // HTTP 409 Conflict: Constraints such as uniqueness are violated.
	DBInvalidCursorState                      = dbError{Code: "24", Message: "Invalid Cursor State"}                        // HTTP 400 Bad Request: The operation on a cursor is invalid.
	DBInvalidTransactionState                 = dbError{Code: "25", Message: "Invalid Transaction State"}                   // HTTP 409 Conflict: Transaction state prevents the operation.
	DBInvalidSQLStatementName                 = dbError{Code: "26", Message: "Invalid SQL Statement Name"}                  // HTTP 400 Bad Request: The SQL statement name is invalid.
	DBTriggeredDataChangeViolation            = dbError{Code: "27", Message: "Triggered Data Change Violation"}             // HTTP 400 Bad Request: Violates database rules triggered by the change.
	DBInvalidAuthorizationSpecification       = dbError{Code: "28", Message: "Invalid Authorization Specification"}         // HTTP 403 Forbidden: User lacks proper authorization.
	DBDependentPrivilegeDescriptorsStillExist = dbError{Code: "2B", Message: "Dependent Privilege Descriptors Still Exist"} // HTTP 400 Bad Request: Cannot remove privileges due to dependencies.
	DBInvalidTransactionTermination           = dbError{Code: "2D", Message: "Invalid Transaction Termination"}             // HTTP 409 Conflict: Transaction termination is invalid in this state.
	DBSQLRoutineException                     = dbError{Code: "2F", Message: "SQL Routine Exception"}                       // HTTP 400 Bad Request: SQL routine encountered a problem.
	DBInvalidCursorName                       = dbError{Code: "34", Message: "Invalid Cursor Name"}                         // HTTP 400 Bad Request: The cursor name does not exist or is invalid.
	DBInvalidCatalogName                      = dbError{Code: "3D", Message: "Invalid Catalog Name"}                        // HTTP 404 Not Found: Specified catalog does not exist.
	DBInvalidSchemaName                       = dbError{Code: "3F", Message: "Invalid Schema Name"}                         // HTTP 404 Not Found: Specified schema does not exist.
	DBSyntaxErrororAccessRuleViolation        = dbError{Code: "42", Message: "Syntax Error or Access Rule Violation"}       // HTTP 400 Bad Request: SQL syntax or access rules are invalid.
	DBWithCheckPointOptionViolation           = dbError{Code: "44", Message: "WITH CHECK OPTION Violation"}                 // HTTP 400 Bad Request: WITH CHECK OPTION constraints violated.

	// Server-side errors (500 range)
	DBConnectionException                = dbError{Code: "08", Message: "Connection Exception"}       // HTTP 500 Internal Server Error: Connection issues between app and database.
	DBTriggeredActionException           = dbError{Code: "09", Message: "Triggered Action Exception"} // HTTP 500 Internal Server Error: Triggered action caused an error.
	DBOperatorIntervention               = dbError{Code: "57", Message: "Operator Intervention"}      // HTTP 500 Internal Server Error: Requires manual intervention to resolve.
	DBSystemError                        = dbError{Code: "58", Message: "System Error"}               // HTTP 500 Internal Server Error: Generic system-level error.
	DBSnapshotFailure                    = dbError{Code: "72", Message: "Snapshot Failure"}           // HTTP 500 Internal Server Error: Snapshot operation failed.
	DBConfigurationFileError             = dbError{Code: "F0", Message: "Configuration File Error"}   // HTTP 500 Internal Server Error: Issues with configuration files.
	DBForeignDataWrapperError            = dbError{Code: "HV", Message: "Foreign Data Wrapper Error"} // HTTP 500 Internal Server Error: Error with foreign data integration.
	DBPLpgSQLError                       = dbError{Code: "P0", Message: "PL/pgSQL Error"}             // HTTP 500 Internal Server Error: Error in procedural code.
	DBInternalError                      = dbError{Code: "XX", Message: "Internal Error"}             // HTTP 500 Internal Server Error: A generic internal error occurred.
	DBGenericError                       = dbError{Code: "500", Message: "DB Error"}                  // HTTP 500 Internal Server Error: Generic database error response.
	DBFeatureNotSupported                = dbError{Code: "0A", Message: "Feature Not Supported"}      // HTTP 501 Not Implemented: The requested operation is not supported by the system.
	DBInvalidTransactionInitiation       = dbError{Code: "0B", Message: "Invalid Transaction Initiation"}
	DBLocatorException                   = dbError{Code: "0F", Message: "Locator Exception"}
	DBDiagnosticsException               = dbError{Code: "0Z", Message: "Diagnostics Exception"}
	DBExternalRoutineException           = dbError{Code: "38", Message: "External Routine Exception"}
	DBExternalRoutineInvocationException = dbError{Code: "39", Message: "External Routine Invocation Exception"}
	DBSavepointException                 = dbError{Code: "3B", Message: "Savepoint Exception"}
	DBTransactionRollback                = dbError{Code: "40", Message: "Transaction Rollback"}
	DBInsufficientResources              = dbError{Code: "53", Message: "Insufficient Resources"}
	DBProgramLimitExceeded               = dbError{Code: "54", Message: "Program Limit Exceeded"}
	DBObjectNotInPrerequisiteState       = dbError{Code: "55", Message: "Object Not In Prerequisite State"}
)
