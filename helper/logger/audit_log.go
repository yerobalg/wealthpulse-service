package logger

type AuditLogEventType string
type AuditLogEventCategory string
type AuditLogEventAction string
type AuditLogEventOutcome string

const (
	AccessType       AuditLogEventType = "access"
	AdminType        AuditLogEventType = "admin"
	AllowedType      AuditLogEventType = "allowed"
	ChangeType       AuditLogEventType = "change"
	ConnectionType   AuditLogEventType = "connection"
	CreationType     AuditLogEventType = "creation"
	DeletionType     AuditLogEventType = "deletion"
	DeniedType       AuditLogEventType = "denied"
	EndType          AuditLogEventType = "end"
	ErrorType        AuditLogEventType = "error"
	GroupType        AuditLogEventType = "group"
	IndicatorType    AuditLogEventType = "indicator"
	InfoType         AuditLogEventType = "info"
	InstallationType AuditLogEventType = "installation"
	ProtocolType     AuditLogEventType = "protocol"
	StartType        AuditLogEventType = "start"
	UserType         AuditLogEventType = "user"

	APICategory                AuditLogEventCategory = "api"
	AuthCategory               AuditLogEventCategory = "authentication"
	ConfigCategory             AuditLogEventCategory = "configuration"
	DatabaseCategory           AuditLogEventCategory = "database"
	DriverCategory             AuditLogEventCategory = "driver"
	EmailCategory              AuditLogEventCategory = "email"
	FileCategory               AuditLogEventCategory = "file"
	HostCategory               AuditLogEventCategory = "host"
	IAMCategory                AuditLogEventCategory = "iam"
	IntrusionDetectionCategory AuditLogEventCategory = "intrusion_detection"
	LibraryCategory            AuditLogEventCategory = "library"
	MalwareCategory            AuditLogEventCategory = "malware"
	NetworkCategory            AuditLogEventCategory = "network"
	PackageCategory            AuditLogEventCategory = "package"
	ProcessCategory            AuditLogEventCategory = "process"
	RegistryCategory           AuditLogEventCategory = "registry"
	SessionCategory            AuditLogEventCategory = "session"
	ThreatCategory             AuditLogEventCategory = "threat"
	VulnerabilityCategory      AuditLogEventCategory = "vulnerability"
	WebCategory                AuditLogEventCategory = "web"

	// Authentication Events

	// Failed login attempt (potential brute force or credential stuffing).
	LoginFailedEventAction AuditLogEventAction = "authn_login_fail"
	// Successful login.
	LoginSuccessEventAction AuditLogEventAction = "authn_login_success"
	// Successful login after previous failures (could indicate account compromise).
	LoginSuccessAfterFailEventAction AuditLogEventAction = "authn_login_successafterfail"
	// Account locked due to security reasons.
	LoginLockEventAction AuditLogEventAction = "authn_login_lock"
	// Maximum failed login attempts reached.
	LoginFailMaxEventAction AuditLogEventAction = "authn_login_fail_max"
	// Password successfully changed.
	PasswordChangeEventAction AuditLogEventAction = "authn_password_change"
	// Failed password change attempt.
	PasswordChangeFailedEventAction AuditLogEventAction = "authn_password_change_fail"
	// Login from geographically impossible locations.
	ImpossibleTravelEventAction AuditLogEventAction = "authn_impossible_travel"
	// Authentication token created.
	TokenCreatedEventAction AuditLogEventAction = "authn_token_created"
	// Authentication token revoked.
	TokenRevokedEventAction AuditLogEventAction = "authn_token_revoked"
	// Authentication token reused.
	TokenReuseEventAction AuditLogEventAction = "authn_token_reuse"
	// Authentication token deleted.
	TokenDeleteEventAction AuditLogEventAction = "authn_token_delete"

	// Authorization Events

	// Unauthorized access attempt.
	AuthorizationFailedEventAction AuditLogEventAction = "authz_fail"
	// User privilege or role changed.
	AuthorizationChangeEventAction AuditLogEventAction = "authz_change"
	// Privileged (admin) action performed.
	AuthorizationAdminEventAction AuditLogEventAction = "authz_admin"

	// Encryption/Decryption Events

	// Decryption operation failed.
	DecryptFailEventAction AuditLogEventAction = "crypt_decrypt_fail"
	// Encryption operation failed.
	EncryptFailEventAction AuditLogEventAction = "crypt_encrypt_fail"

	// Resource Usage Events

	// User or system exceeded allowed usage limits.
	ExcessRateLimitExceededEventAction AuditLogEventAction = "excess_rate_limit_exceeded"

	// File Upload Events

	// File successfully uploaded.
	UploadCompleteEventAction AuditLogEventAction = "upload_complete"
	// Uploaded file stored.
	UploadStoredEventAction AuditLogEventAction = "upload_stored"
	// File validation (e.g., virus scan, type check) result.
	UploadValidationEventAction AuditLogEventAction = "upload_validation"
	// Uploaded file deleted.
	UploadDeleteEventAction AuditLogEventAction = "upload_delete"

	// Input Validation Events

	// Server-side input validation failed (possible attack attempt).
	InputValidationFailEventAction AuditLogEventAction = "input_validation_fail"

	// Malicious Behavior Detection

	// Excessive 404 errors (possible probing).
	MaliciousExcess404EventAction AuditLogEventAction = "malicious_excess_404"
	// Unexpected or extraneous input detected.
	MaliciousExtraneousEventAction AuditLogEventAction = "malicious_extraneous"
	// Known attack tool detected.
	MaliciousAttackToolEventAction AuditLogEventAction = "malicious_attack_tool"
	// Unauthorized cross-origin request detected.
	MaliciousCORSDetectedEventAction AuditLogEventAction = "malicious_cors"
	// Direct object reference attempt.
	MaliciousDirectReferenceEventAction AuditLogEventAction = "malicious_direct_reference"

	// Privilege Management Events

	// Access control/permissions for an object changed.
	PrivilegePermissionsChangedEventAction AuditLogEventAction = "privilege_permissions_changed"

	// Sensitive Data Access Events

	// Sensitive data created.
	SensitiveCreateEventAction AuditLogEventAction = "sensitive_create"
	// Sensitive data accessed.
	SensitiveReadEventAction AuditLogEventAction = "sensitive_read"
	// Sensitive data modified.
	SensitiveUpdateEventAction AuditLogEventAction = "sensitive_update"
	// Sensitive data deleted.
	SensitiveDeleteEventAction AuditLogEventAction = "sensitive_delete"

	// Application Flow Events

	// Application sequence/logic violated.
	SequenceFailEventAction AuditLogEventAction = "sequence_fail"

	// Session Management Events

	// Session created.
	SessionCreatedEventAction AuditLogEventAction = "session_created"
	// Session renewed.
	SessionRenewedEventAction AuditLogEventAction = "session_renewed"
	// Session expired.
	SessionExpiredEventAction AuditLogEventAction = "session_expired"
	// Attempt to use expired session.
	SessionUseAfterExpireEventAction AuditLogEventAction = "session_use_after_expire"

	// System Events

	// System startup.
	SysStartupEventAction AuditLogEventAction = "sys_startup"
	// System shutdown.
	SysShutdownEventAction AuditLogEventAction = "sys_shutdown"
	// System restarted.
	SysRestartEventAction AuditLogEventAction = "sys_restart"
	// System crashed.
	SysCrashEventAction AuditLogEventAction = "sys_crash"
	// Monitoring disabled.
	SysMonitorDisabledEventAction AuditLogEventAction = "sys_monitor_disabled"
	// Monitoring enabled.
	SysMonitorEnabledEventAction AuditLogEventAction = "sys_monitor_enabled"

	// User Management Events

	// User account created.
	UserCreatedEventAction AuditLogEventAction = "user_created"
	// User account updated.
	UserUpdatedEventAction AuditLogEventAction = "user_updated"
	// User account archived.
	UserArchivedEventAction AuditLogEventAction = "user_archived"
	// User account deleted.
	UserDeletedEventAction AuditLogEventAction = "user_deleted"

	SuccessEventOutcome AuditLogEventOutcome = "success"
	FailureEventOutcome AuditLogEventOutcome = "failure"
)

type AuditLogEvent struct {
	Type     []AuditLogEventType     `json:"type"`
	Category []AuditLogEventCategory `json:"category"`
	Action   AuditLogEventAction     `json:"action"`
	Outcome  AuditLogEventOutcome    `json:"outcome"`
}
