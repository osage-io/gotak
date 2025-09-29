// Package rbac provides Role-Based Access Control (RBAC) and Attribute-Based Access Control (ABAC)
// implementations for fine-grained authorization in the GoTAK tactical awareness platform.
package rbac

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AccessControlManager defines the main interface for RBAC/ABAC operations
type AccessControlManager interface {
	// Role management
	CreateRole(ctx context.Context, role *Role) error
	GetRole(ctx context.Context, roleID uuid.UUID) (*Role, error)
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, roleID uuid.UUID) error
	ListRoles(ctx context.Context) ([]*Role, error)

	// Permission management
	CreatePermission(ctx context.Context, permission *Permission) error
	GetPermission(ctx context.Context, permissionID uuid.UUID) (*Permission, error)
	UpdatePermission(ctx context.Context, permission *Permission) error
	DeletePermission(ctx context.Context, permissionID uuid.UUID) error
	ListPermissions(ctx context.Context) ([]*Permission, error)

	// Role assignments
	AssignRoleToUser(ctx context.Context, userID uuid.UUID, roleID uuid.UUID, metadata map[string]string) error
	RevokeRoleFromUser(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*RoleBinding, error)
	GetRoleUsers(ctx context.Context, roleID uuid.UUID) ([]*RoleBinding, error)

	// Authorization decisions
	CheckPermission(ctx context.Context, request *AuthorizationRequest) (*AuthorizationDecision, error)
	BatchCheckPermissions(ctx context.Context, requests []*AuthorizationRequest) ([]*AuthorizationDecision, error)

	// Policy evaluation (ABAC)
	EvaluatePolicy(ctx context.Context, policy *Policy, attributes map[string]interface{}) (*PolicyDecision, error)
	CreatePolicy(ctx context.Context, policy *Policy) error
	UpdatePolicy(ctx context.Context, policy *Policy) error
	DeletePolicy(ctx context.Context, policyID uuid.UUID) error
	ListPolicies(ctx context.Context) ([]*Policy, error)
}

// Role represents a role in the RBAC system
type Role struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	Name        string      `json:"name" db:"name"`                               // e.g., "tactical:commander"
	DisplayName string      `json:"display_name" db:"display_name"`               // e.g., "Tactical Commander"
	Description string      `json:"description" db:"description"`                 // Human-readable description
	Type        RoleType    `json:"type" db:"type"`                               // system, tactical, custom
	ParentID    *uuid.UUID  `json:"parent_id,omitempty" db:"parent_id"`           // For role hierarchy
	Permissions []uuid.UUID `json:"permissions" db:"permissions"`                 // Associated permission IDs
	Metadata    Metadata    `json:"metadata" db:"metadata"`                       // Additional role data
	IsActive    bool        `json:"is_active" db:"is_active"`                     // Whether role is active
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
	CreatedBy   uuid.UUID   `json:"created_by" db:"created_by"`                   // User who created the role
}

// Permission represents a specific permission in the system
type Permission struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	Name        string         `json:"name" db:"name"`                     // e.g., "missions:read"
	DisplayName string         `json:"display_name" db:"display_name"`     // e.g., "Read Missions"
	Description string         `json:"description" db:"description"`       // Human-readable description
	Resource    string         `json:"resource" db:"resource"`             // Resource type (missions, users, etc.)
	Action      string         `json:"action" db:"action"`                 // Action (read, write, delete, etc.)
	Effect      PermissionType `json:"effect" db:"effect"`                 // allow, deny
	Conditions  []Condition    `json:"conditions,omitempty" db:"conditions"` // Optional conditions
	Metadata    Metadata       `json:"metadata" db:"metadata"`
	IsActive    bool           `json:"is_active" db:"is_active"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
	CreatedBy   uuid.UUID      `json:"created_by" db:"created_by"`
}

// RoleBinding represents the assignment of a role to a user
type RoleBinding struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	RoleID     uuid.UUID `json:"role_id" db:"role_id"`
	Role       *Role     `json:"role,omitempty"`                         // Populated in queries
	GrantedBy  uuid.UUID `json:"granted_by" db:"granted_by"`             // User who granted the role
	Metadata   Metadata  `json:"metadata" db:"metadata"`                 // Additional binding data
	ExpiresAt  *time.Time `json:"expires_at,omitempty" db:"expires_at"`  // Optional expiration
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// Policy represents an ABAC policy for attribute-based access control
type Policy struct {
	ID          uuid.UUID    `json:"id" db:"id"`
	Name        string       `json:"name" db:"name"`                     // e.g., "time_restricted_access"
	DisplayName string       `json:"display_name" db:"display_name"`     // e.g., "Time-Restricted Access"
	Description string       `json:"description" db:"description"`
	Type        PolicyType   `json:"type" db:"type"`                     // rbac, abac, hybrid
	Rules       []PolicyRule `json:"rules" db:"rules"`                   // Policy rules
	Effect      EffectType   `json:"effect" db:"effect"`                 // allow, deny
	Priority    int          `json:"priority" db:"priority"`             // Higher number = higher priority
	IsActive    bool         `json:"is_active" db:"is_active"`
	Version     int          `json:"version" db:"version"`               // For policy versioning
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	CreatedBy   uuid.UUID    `json:"created_by" db:"created_by"`
}

// PolicyRule represents a single rule within a policy
type PolicyRule struct {
	ID          string                 `json:"id"`                                    // Unique rule ID within policy
	Name        string                 `json:"name"`                                  // Human-readable name
	Condition   string                 `json:"condition"`                             // JSON Logic expression
	Effect      EffectType             `json:"effect"`                                // allow, deny
	Attributes  map[string]interface{} `json:"attributes,omitempty"`                  // Required attributes
	Resources   []string               `json:"resources,omitempty"`                   // Applicable resources
	Actions     []string               `json:"actions,omitempty"`                     // Applicable actions
	Context     map[string]interface{} `json:"context,omitempty"`                     // Contextual constraints
}

// Condition represents a conditional constraint on a permission or policy
type Condition struct {
	Type      ConditionType          `json:"type"`      // time, location, network, attribute, etc.
	Field     string                 `json:"field"`     // Field to evaluate
	Operator  string                 `json:"operator"`  // eq, ne, gt, lt, contains, etc.
	Value     interface{}            `json:"value"`     // Expected value
	Context   map[string]interface{} `json:"context,omitempty"` // Additional context
}

// AuthorizationRequest represents a request for authorization
type AuthorizationRequest struct {
	UserID     uuid.UUID              `json:"user_id"`                        // User making the request
	Resource   string                 `json:"resource"`                       // Resource being accessed
	Action     string                 `json:"action"`                         // Action being performed
	Context    map[string]interface{} `json:"context,omitempty"`              // Request context
	Attributes map[string]interface{} `json:"attributes,omitempty"`           // User/session attributes
	IPAddress  string                 `json:"ip_address,omitempty"`           // Client IP address
	UserAgent  string                 `json:"user_agent,omitempty"`           // Client user agent
	Timestamp  time.Time              `json:"timestamp"`                      // Request timestamp
}

// AuthorizationDecision represents the result of an authorization check
type AuthorizationDecision struct {
	RequestID    string        `json:"request_id"`                        // Unique request identifier
	UserID       uuid.UUID     `json:"user_id"`
	Resource     string        `json:"resource"`
	Action       string        `json:"action"`
	Decision     DecisionType  `json:"decision"`                          // allow, deny, indeterminate
	Reason       string        `json:"reason"`                            // Human-readable reason
	AppliedRoles []uuid.UUID   `json:"applied_roles,omitempty"`           // Roles that influenced decision
	AppliedPolicies []uuid.UUID `json:"applied_policies,omitempty"`       // Policies that influenced decision
	Evidence     []Evidence    `json:"evidence,omitempty"`                // Supporting evidence
	TTL          time.Duration `json:"ttl,omitempty"`                     // Cache TTL for decision
	Metadata     Metadata      `json:"metadata,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
}

// PolicyDecision represents the result of a policy evaluation
type PolicyDecision struct {
	PolicyID    uuid.UUID              `json:"policy_id"`
	Decision    DecisionType           `json:"decision"`              // allow, deny, indeterminate
	Reason      string                 `json:"reason"`
	MatchedRules []string              `json:"matched_rules,omitempty"` // Rule IDs that matched
	Context     map[string]interface{} `json:"context,omitempty"`     // Evaluation context
	Timestamp   time.Time              `json:"timestamp"`
}

// Evidence represents supporting evidence for an authorization decision
type Evidence struct {
	Type        EvidenceType           `json:"type"`        // role, permission, policy, condition
	Source      string                 `json:"source"`      // Source identifier
	Value       interface{}            `json:"value"`       // Evidence value
	Context     map[string]interface{} `json:"context,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// Metadata represents flexible key-value metadata
type Metadata map[string]interface{}

// Enum types

// RoleType represents the type of role
type RoleType string

const (
	RoleTypeSystem   RoleType = "system"   // System-defined roles
	RoleTypeTactical RoleType = "tactical" // Tactical/military roles
	RoleTypeCustom   RoleType = "custom"   // User-defined roles
)

// PermissionType represents the effect of a permission
type PermissionType string

const (
	PermissionTypeAllow PermissionType = "allow"
	PermissionTypeDeny  PermissionType = "deny"
)

// PolicyType represents the type of policy
type PolicyType string

const (
	PolicyTypeRBAC   PolicyType = "rbac"   // Role-based policy
	PolicyTypeABAC   PolicyType = "abac"   // Attribute-based policy
	PolicyTypeHybrid PolicyType = "hybrid" // Hybrid RBAC/ABAC policy
)

// EffectType represents the effect of a policy or rule
type EffectType string

const (
	EffectTypeAllow EffectType = "allow"
	EffectTypeDeny  EffectType = "deny"
)

// DecisionType represents an authorization decision
type DecisionType string

const (
	DecisionTypeAllow         DecisionType = "allow"
	DecisionTypeDeny          DecisionType = "deny"
	DecisionTypeIndeterminate DecisionType = "indeterminate" // Cannot determine
)

// ConditionType represents the type of condition
type ConditionType string

const (
	ConditionTypeTime      ConditionType = "time"      // Time-based conditions
	ConditionTypeLocation  ConditionType = "location"  // Geographic conditions
	ConditionTypeNetwork   ConditionType = "network"   // Network-based conditions
	ConditionTypeAttribute ConditionType = "attribute" // Attribute-based conditions
	ConditionTypeCustom    ConditionType = "custom"    // Custom conditions
)

// EvidenceType represents the type of evidence
type EvidenceType string

const (
	EvidenceTypeRole       EvidenceType = "role"
	EvidenceTypePermission EvidenceType = "permission"
	EvidenceTypePolicy     EvidenceType = "policy"
	EvidenceTypeCondition  EvidenceType = "condition"
)

// Repository interfaces

// RoleRepository manages role persistence
type RoleRepository interface {
	CreateRole(ctx context.Context, role *Role) error
	GetRole(ctx context.Context, roleID uuid.UUID) (*Role, error)
	GetRoleByName(ctx context.Context, name string) (*Role, error)
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, roleID uuid.UUID) error
	ListRoles(ctx context.Context, filters RoleFilters) ([]*Role, int, error)
	GetRoleHierarchy(ctx context.Context, roleID uuid.UUID) ([]*Role, error)
}

// PermissionRepository manages permission persistence
type PermissionRepository interface {
	CreatePermission(ctx context.Context, permission *Permission) error
	GetPermission(ctx context.Context, permissionID uuid.UUID) (*Permission, error)
	GetPermissionByName(ctx context.Context, name string) (*Permission, error)
	UpdatePermission(ctx context.Context, permission *Permission) error
	DeletePermission(ctx context.Context, permissionID uuid.UUID) error
	ListPermissions(ctx context.Context, filters PermissionFilters) ([]*Permission, int, error)
	GetPermissionsByResource(ctx context.Context, resource string) ([]*Permission, error)
}

// RoleBindingRepository manages role assignments
type RoleBindingRepository interface {
	CreateRoleBinding(ctx context.Context, binding *RoleBinding) error
	GetRoleBinding(ctx context.Context, bindingID uuid.UUID) (*RoleBinding, error)
	GetUserRoleBindings(ctx context.Context, userID uuid.UUID) ([]*RoleBinding, error)
	GetRoleBindings(ctx context.Context, roleID uuid.UUID) ([]*RoleBinding, error)
	UpdateRoleBinding(ctx context.Context, binding *RoleBinding) error
	DeleteRoleBinding(ctx context.Context, bindingID uuid.UUID) error
	DeleteUserRoleBinding(ctx context.Context, userID, roleID uuid.UUID) error
}

// PolicyRepository manages policy persistence
type PolicyRepository interface {
	CreatePolicy(ctx context.Context, policy *Policy) error
	GetPolicy(ctx context.Context, policyID uuid.UUID) (*Policy, error)
	GetPolicyByName(ctx context.Context, name string) (*Policy, error)
	UpdatePolicy(ctx context.Context, policy *Policy) error
	DeletePolicy(ctx context.Context, policyID uuid.UUID) error
	ListPolicies(ctx context.Context, filters PolicyFilters) ([]*Policy, int, error)
	GetActivePolicies(ctx context.Context) ([]*Policy, error)
}

// Filter types for queries

// RoleFilters defines filters for role queries
type RoleFilters struct {
	Type      *RoleType `json:"type,omitempty"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	IsActive  *bool     `json:"is_active,omitempty"`
	Limit     int       `json:"limit,omitempty"`
	Offset    int       `json:"offset,omitempty"`
}

// PermissionFilters defines filters for permission queries
type PermissionFilters struct {
	Resource  string           `json:"resource,omitempty"`
	Action    string           `json:"action,omitempty"`
	Effect    *PermissionType  `json:"effect,omitempty"`
	IsActive  *bool            `json:"is_active,omitempty"`
	Limit     int              `json:"limit,omitempty"`
	Offset    int              `json:"offset,omitempty"`
}

// PolicyFilters defines filters for policy queries
type PolicyFilters struct {
	Type      *PolicyType `json:"type,omitempty"`
	IsActive  *bool       `json:"is_active,omitempty"`
	Priority  *int        `json:"priority,omitempty"`
	Limit     int         `json:"limit,omitempty"`
	Offset    int         `json:"offset,omitempty"`
}

// AttributeEvaluator handles ABAC attribute evaluation
type AttributeEvaluator interface {
	// Evaluate evaluates a JSON Logic expression against provided attributes
	Evaluate(expression string, attributes map[string]interface{}) (bool, error)
	
	// ValidateExpression validates that a JSON Logic expression is valid
	ValidateExpression(expression string) error
	
	// ExtractRequiredAttributes extracts the required attributes from an expression
	ExtractRequiredAttributes(expression string) ([]string, error)
}

// AccessControlError represents access control related errors
type AccessControlError struct {
	Type    string            `json:"type"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
	Cause   error             `json:"-"`
}

func (e *AccessControlError) Error() string {
	return e.Message
}

func (e *AccessControlError) Unwrap() error {
	return e.Cause
}

// Common access control error types
const (
	ErrTypeRoleNotFound       = "role_not_found"
	ErrTypePermissionNotFound = "permission_not_found"
	ErrTypePolicyNotFound     = "policy_not_found"
	ErrTypeInvalidRole        = "invalid_role"
	ErrTypeInvalidPermission  = "invalid_permission"
	ErrTypeInvalidPolicy      = "invalid_policy"
	ErrTypeAccessDenied       = "access_denied"
	ErrTypeInvalidRequest     = "invalid_request"
	ErrTypePolicyEvaluation   = "policy_evaluation_error"
	ErrTypeStorageError       = "storage_error"
)

// NewAccessControlError creates a new access control error
func NewAccessControlError(errorType, message string, cause error) *AccessControlError {
	return &AccessControlError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
	}
}

// Helper functions for working with roles and permissions

// HasPermission checks if a set of roles has a specific permission
func HasPermission(roles []*Role, resource, action string) bool {
	for _, role := range roles {
		for _, permID := range role.Permissions {
			// This would normally look up the permission from storage
			// For now, we'll assume the caller has resolved permissions
			_ = permID
		}
	}
	return false
}

// IsRoleHierarchyValid checks if a role hierarchy is valid (no cycles)
func IsRoleHierarchyValid(role *Role, allRoles []*Role) bool {
	visited := make(map[uuid.UUID]bool)
	return !hasCycle(role, allRoles, visited)
}

func hasCycle(role *Role, allRoles []*Role, visited map[uuid.UUID]bool) bool {
	if visited[role.ID] {
		return true // Cycle detected
	}
	
	if role.ParentID == nil {
		return false // No parent, no cycle
	}
	
	visited[role.ID] = true
	
	// Find parent role
	for _, r := range allRoles {
		if r.ID == *role.ParentID {
			return hasCycle(r, allRoles, visited)
		}
	}
	
	return false // Parent not found, no cycle
}
