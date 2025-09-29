package rbac

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// rbacManager implements the AccessControlManager interface
type rbacManager struct {
	roleRepo        RoleRepository
	permissionRepo  PermissionRepository
	roleBindingRepo RoleBindingRepository
	policyRepo      PolicyRepository
	attributeEval   AttributeEvaluator
	logger          *logrus.Logger
}

// NewAccessControlManager creates a new RBAC/ABAC manager
func NewAccessControlManager(
	roleRepo RoleRepository,
	permissionRepo PermissionRepository,
	roleBindingRepo RoleBindingRepository,
	policyRepo PolicyRepository,
	attributeEval AttributeEvaluator,
	logger *logrus.Logger,
) AccessControlManager {
	return &rbacManager{
		roleRepo:        roleRepo,
		permissionRepo:  permissionRepo,
		roleBindingRepo: roleBindingRepo,
		policyRepo:      policyRepo,
		attributeEval:   attributeEval,
		logger:          logger,
	}
}

// Role management

func (m *rbacManager) CreateRole(ctx context.Context, role *Role) error {
	if role == nil {
		return NewAccessControlError(ErrTypeInvalidRole, "role cannot be nil", nil)
	}

	if err := m.validateRole(role); err != nil {
		return err
	}

	// Generate ID and timestamps
	role.ID = uuid.New()
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	if err := m.roleRepo.CreateRole(ctx, role); err != nil {
		m.logger.WithError(err).WithField("role_name", role.Name).Error("Failed to create role")
		return NewAccessControlError(ErrTypeStorageError, "failed to create role", err)
	}

	m.logger.WithFields(logrus.Fields{
		"role_id":   role.ID,
		"role_name": role.Name,
		"role_type": role.Type,
	}).Info("Role created successfully")

	return nil
}

func (m *rbacManager) GetRole(ctx context.Context, roleID uuid.UUID) (*Role, error) {
	role, err := m.roleRepo.GetRole(ctx, roleID)
	if err != nil {
		m.logger.WithError(err).WithField("role_id", roleID).Error("Failed to get role")
		return nil, NewAccessControlError(ErrTypeRoleNotFound, "role not found", err)
	}

	return role, nil
}

func (m *rbacManager) UpdateRole(ctx context.Context, role *Role) error {
	if role == nil {
		return NewAccessControlError(ErrTypeInvalidRole, "role cannot be nil", nil)
	}

	if err := m.validateRole(role); err != nil {
		return err
	}

	// Check if role exists
	existing, err := m.roleRepo.GetRole(ctx, role.ID)
	if err != nil {
		return NewAccessControlError(ErrTypeRoleNotFound, "role not found", err)
	}

	// Preserve creation info
	role.CreatedAt = existing.CreatedAt
	role.CreatedBy = existing.CreatedBy
	role.UpdatedAt = time.Now()

	if err := m.roleRepo.UpdateRole(ctx, role); err != nil {
		m.logger.WithError(err).WithField("role_id", role.ID).Error("Failed to update role")
		return NewAccessControlError(ErrTypeStorageError, "failed to update role", err)
	}

	m.logger.WithField("role_id", role.ID).Info("Role updated successfully")
	return nil
}

func (m *rbacManager) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	// Check if role has any bindings
	bindings, err := m.roleBindingRepo.GetRoleBindings(ctx, roleID)
	if err != nil {
		return NewAccessControlError(ErrTypeStorageError, "failed to check role bindings", err)
	}

	if len(bindings) > 0 {
		return NewAccessControlError(ErrTypeInvalidRequest, "cannot delete role with active bindings", nil)
	}

	if err := m.roleRepo.DeleteRole(ctx, roleID); err != nil {
		m.logger.WithError(err).WithField("role_id", roleID).Error("Failed to delete role")
		return NewAccessControlError(ErrTypeStorageError, "failed to delete role", err)
	}

	m.logger.WithField("role_id", roleID).Info("Role deleted successfully")
	return nil
}

func (m *rbacManager) ListRoles(ctx context.Context) ([]*Role, error) {
	roles, _, err := m.roleRepo.ListRoles(ctx, RoleFilters{})
	if err != nil {
		m.logger.WithError(err).Error("Failed to list roles")
		return nil, NewAccessControlError(ErrTypeStorageError, "failed to list roles", err)
	}

	return roles, nil
}

// Permission management

func (m *rbacManager) CreatePermission(ctx context.Context, permission *Permission) error {
	if permission == nil {
		return NewAccessControlError(ErrTypeInvalidPermission, "permission cannot be nil", nil)
	}

	if err := m.validatePermission(permission); err != nil {
		return err
	}

	permission.ID = uuid.New()
	permission.CreatedAt = time.Now()
	permission.UpdatedAt = time.Now()

	if err := m.permissionRepo.CreatePermission(ctx, permission); err != nil {
		m.logger.WithError(err).WithField("permission_name", permission.Name).Error("Failed to create permission")
		return NewAccessControlError(ErrTypeStorageError, "failed to create permission", err)
	}

	m.logger.WithFields(logrus.Fields{
		"permission_id":   permission.ID,
		"permission_name": permission.Name,
		"resource":        permission.Resource,
		"action":          permission.Action,
	}).Info("Permission created successfully")

	return nil
}

func (m *rbacManager) GetPermission(ctx context.Context, permissionID uuid.UUID) (*Permission, error) {
	permission, err := m.permissionRepo.GetPermission(ctx, permissionID)
	if err != nil {
		m.logger.WithError(err).WithField("permission_id", permissionID).Error("Failed to get permission")
		return nil, NewAccessControlError(ErrTypePermissionNotFound, "permission not found", err)
	}

	return permission, nil
}

func (m *rbacManager) UpdatePermission(ctx context.Context, permission *Permission) error {
	if permission == nil {
		return NewAccessControlError(ErrTypeInvalidPermission, "permission cannot be nil", nil)
	}

	if err := m.validatePermission(permission); err != nil {
		return err
	}

	existing, err := m.permissionRepo.GetPermission(ctx, permission.ID)
	if err != nil {
		return NewAccessControlError(ErrTypePermissionNotFound, "permission not found", err)
	}

	permission.CreatedAt = existing.CreatedAt
	permission.CreatedBy = existing.CreatedBy
	permission.UpdatedAt = time.Now()

	if err := m.permissionRepo.UpdatePermission(ctx, permission); err != nil {
		m.logger.WithError(err).WithField("permission_id", permission.ID).Error("Failed to update permission")
		return NewAccessControlError(ErrTypeStorageError, "failed to update permission", err)
	}

	m.logger.WithField("permission_id", permission.ID).Info("Permission updated successfully")
	return nil
}

func (m *rbacManager) DeletePermission(ctx context.Context, permissionID uuid.UUID) error {
	if err := m.permissionRepo.DeletePermission(ctx, permissionID); err != nil {
		m.logger.WithError(err).WithField("permission_id", permissionID).Error("Failed to delete permission")
		return NewAccessControlError(ErrTypeStorageError, "failed to delete permission", err)
	}

	m.logger.WithField("permission_id", permissionID).Info("Permission deleted successfully")
	return nil
}

func (m *rbacManager) ListPermissions(ctx context.Context) ([]*Permission, error) {
	permissions, _, err := m.permissionRepo.ListPermissions(ctx, PermissionFilters{})
	if err != nil {
		m.logger.WithError(err).Error("Failed to list permissions")
		return nil, NewAccessControlError(ErrTypeStorageError, "failed to list permissions", err)
	}

	return permissions, nil
}

// Role assignment management

func (m *rbacManager) AssignRoleToUser(ctx context.Context, userID uuid.UUID, roleID uuid.UUID, metadata map[string]string) error {
	// Check if role exists
	role, err := m.roleRepo.GetRole(ctx, roleID)
	if err != nil {
		return NewAccessControlError(ErrTypeRoleNotFound, "role not found", err)
	}

	if !role.IsActive {
		return NewAccessControlError(ErrTypeInvalidRequest, "cannot assign inactive role", nil)
	}

	// Check if binding already exists
	existingBindings, err := m.roleBindingRepo.GetUserRoleBindings(ctx, userID)
	if err != nil {
		return NewAccessControlError(ErrTypeStorageError, "failed to check existing bindings", err)
	}

	for _, binding := range existingBindings {
		if binding.RoleID == roleID && binding.IsActive {
			return NewAccessControlError(ErrTypeInvalidRequest, "role already assigned to user", nil)
		}
	}

	// Create new role binding
	binding := &RoleBinding{
		ID:        uuid.New(),
		UserID:    userID,
		RoleID:    roleID,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(Metadata),
	}

	// Add metadata
	for k, v := range metadata {
		binding.Metadata[k] = v
	}

	if err := m.roleBindingRepo.CreateRoleBinding(ctx, binding); err != nil {
		m.logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"role_id": roleID,
		}).Error("Failed to create role binding")
		return NewAccessControlError(ErrTypeStorageError, "failed to assign role", err)
	}

	m.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"role_id":   roleID,
		"role_name": role.Name,
	}).Info("Role assigned to user successfully")

	return nil
}

func (m *rbacManager) RevokeRoleFromUser(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	if err := m.roleBindingRepo.DeleteUserRoleBinding(ctx, userID, roleID); err != nil {
		m.logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"role_id": roleID,
		}).Error("Failed to revoke role from user")
		return NewAccessControlError(ErrTypeStorageError, "failed to revoke role", err)
	}

	m.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"role_id": roleID,
	}).Info("Role revoked from user successfully")

	return nil
}

func (m *rbacManager) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*RoleBinding, error) {
	bindings, err := m.roleBindingRepo.GetUserRoleBindings(ctx, userID)
	if err != nil {
		m.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user roles")
		return nil, NewAccessControlError(ErrTypeStorageError, "failed to get user roles", err)
	}

	return bindings, nil
}

func (m *rbacManager) GetRoleUsers(ctx context.Context, roleID uuid.UUID) ([]*RoleBinding, error) {
	bindings, err := m.roleBindingRepo.GetRoleBindings(ctx, roleID)
	if err != nil {
		m.logger.WithError(err).WithField("role_id", roleID).Error("Failed to get role users")
		return nil, NewAccessControlError(ErrTypeStorageError, "failed to get role users", err)
	}

	return bindings, nil
}

// Authorization decisions

func (m *rbacManager) CheckPermission(ctx context.Context, request *AuthorizationRequest) (*AuthorizationDecision, error) {
	if request == nil {
		return nil, NewAccessControlError(ErrTypeInvalidRequest, "authorization request cannot be nil", nil)
	}

	requestID := uuid.New().String()
	startTime := time.Now()

	m.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"user_id":    request.UserID,
		"resource":   request.Resource,
		"action":     request.Action,
	}).Debug("Processing authorization request")

	// Get user's role bindings
	roleBindings, err := m.roleBindingRepo.GetUserRoleBindings(ctx, request.UserID)
	if err != nil {
		return nil, NewAccessControlError(ErrTypeStorageError, "failed to get user roles", err)
	}

	decision := &AuthorizationDecision{
		RequestID:       requestID,
		UserID:          request.UserID,
		Resource:        request.Resource,
		Action:          request.Action,
		Decision:        DecisionTypeDeny, // Default to deny
		Reason:          "No matching permissions found",
		AppliedRoles:    []uuid.UUID{},
		AppliedPolicies: []uuid.UUID{},
		Evidence:        []Evidence{},
		Timestamp:       startTime,
	}

	// Check RBAC permissions
	allowed, evidence, err := m.checkRBACPermissions(ctx, request, roleBindings)
	if err != nil {
		decision.Reason = fmt.Sprintf("RBAC evaluation error: %v", err)
		return decision, nil
	}

	if allowed {
		decision.Decision = DecisionTypeAllow
		decision.Reason = "Access granted via RBAC permissions"
		decision.Evidence = evidence
		for _, binding := range roleBindings {
			if binding.IsActive {
				decision.AppliedRoles = append(decision.AppliedRoles, binding.RoleID)
			}
		}
		return decision, nil
	}

	// Check ABAC policies if RBAC doesn't allow
	policyDecision, err := m.checkABACPolicies(ctx, request)
	if err != nil {
		decision.Reason = fmt.Sprintf("ABAC evaluation error: %v", err)
		return decision, nil
	}

	if policyDecision != nil && policyDecision.Decision == DecisionTypeAllow {
		decision.Decision = DecisionTypeAllow
		decision.Reason = "Access granted via ABAC policies"
		decision.AppliedPolicies = append(decision.AppliedPolicies, policyDecision.PolicyID)
		
		// Add policy evidence
		decision.Evidence = append(decision.Evidence, Evidence{
			Type:      EvidenceTypePolicy,
			Source:    policyDecision.PolicyID.String(),
			Value:     policyDecision.MatchedRules,
			Context:   policyDecision.Context,
			Timestamp: time.Now(),
		})
	}

	m.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"decision":   decision.Decision,
		"reason":     decision.Reason,
		"duration":   time.Since(startTime),
	}).Info("Authorization decision made")

	return decision, nil
}

func (m *rbacManager) BatchCheckPermissions(ctx context.Context, requests []*AuthorizationRequest) ([]*AuthorizationDecision, error) {
	if len(requests) == 0 {
		return []*AuthorizationDecision{}, nil
	}

	decisions := make([]*AuthorizationDecision, len(requests))
	for i, request := range requests {
		decision, err := m.CheckPermission(ctx, request)
		if err != nil {
			// Create a deny decision with error information
			decisions[i] = &AuthorizationDecision{
				RequestID: uuid.New().String(),
				UserID:    request.UserID,
				Resource:  request.Resource,
				Action:    request.Action,
				Decision:  DecisionTypeDeny,
				Reason:    fmt.Sprintf("Authorization error: %v", err),
				Timestamp: time.Now(),
			}
		} else {
			decisions[i] = decision
		}
	}

	return decisions, nil
}

// Policy management and evaluation

func (m *rbacManager) EvaluatePolicy(ctx context.Context, policy *Policy, attributes map[string]interface{}) (*PolicyDecision, error) {
	if policy == nil {
		return nil, NewAccessControlError(ErrTypeInvalidPolicy, "policy cannot be nil", nil)
	}

	decision := &PolicyDecision{
		PolicyID:     policy.ID,
		Decision:     DecisionTypeDeny,
		Reason:       "No matching rules",
		MatchedRules: []string{},
		Context:      make(map[string]interface{}),
		Timestamp:    time.Now(),
	}

	if !policy.IsActive {
		decision.Reason = "Policy is not active"
		return decision, nil
	}

	// Evaluate policy rules
	for _, rule := range policy.Rules {
		matches, err := m.attributeEval.Evaluate(rule.Condition, attributes)
		if err != nil {
			m.logger.WithError(err).WithFields(logrus.Fields{
				"policy_id": policy.ID,
				"rule_id":   rule.ID,
			}).Error("Failed to evaluate policy rule")
			continue
		}

		if matches {
			decision.MatchedRules = append(decision.MatchedRules, rule.ID)
			decision.Context[fmt.Sprintf("rule_%s", rule.ID)] = rule.Context

			if rule.Effect == EffectTypeAllow {
				decision.Decision = DecisionTypeAllow
				decision.Reason = fmt.Sprintf("Rule %s (%s) matched and allows access", rule.ID, rule.Name)
			} else if rule.Effect == EffectTypeDeny {
				decision.Decision = DecisionTypeDeny
				decision.Reason = fmt.Sprintf("Rule %s (%s) matched and denies access", rule.ID, rule.Name)
				// Deny rules take precedence, so we can return early
				return decision, nil
			}
		}
	}

	return decision, nil
}

func (m *rbacManager) CreatePolicy(ctx context.Context, policy *Policy) error {
	if policy == nil {
		return NewAccessControlError(ErrTypeInvalidPolicy, "policy cannot be nil", nil)
	}

	if err := m.validatePolicy(policy); err != nil {
		return err
	}

	policy.ID = uuid.New()
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()
	policy.Version = 1

	if err := m.policyRepo.CreatePolicy(ctx, policy); err != nil {
		m.logger.WithError(err).WithField("policy_name", policy.Name).Error("Failed to create policy")
		return NewAccessControlError(ErrTypeStorageError, "failed to create policy", err)
	}

	m.logger.WithFields(logrus.Fields{
		"policy_id":   policy.ID,
		"policy_name": policy.Name,
		"policy_type": policy.Type,
	}).Info("Policy created successfully")

	return nil
}

func (m *rbacManager) UpdatePolicy(ctx context.Context, policy *Policy) error {
	if policy == nil {
		return NewAccessControlError(ErrTypeInvalidPolicy, "policy cannot be nil", nil)
	}

	if err := m.validatePolicy(policy); err != nil {
		return err
	}

	existing, err := m.policyRepo.GetPolicy(ctx, policy.ID)
	if err != nil {
		return NewAccessControlError(ErrTypePolicyNotFound, "policy not found", err)
	}

	policy.CreatedAt = existing.CreatedAt
	policy.CreatedBy = existing.CreatedBy
	policy.UpdatedAt = time.Now()
	policy.Version = existing.Version + 1

	if err := m.policyRepo.UpdatePolicy(ctx, policy); err != nil {
		m.logger.WithError(err).WithField("policy_id", policy.ID).Error("Failed to update policy")
		return NewAccessControlError(ErrTypeStorageError, "failed to update policy", err)
	}

	m.logger.WithField("policy_id", policy.ID).Info("Policy updated successfully")
	return nil
}

func (m *rbacManager) DeletePolicy(ctx context.Context, policyID uuid.UUID) error {
	if err := m.policyRepo.DeletePolicy(ctx, policyID); err != nil {
		m.logger.WithError(err).WithField("policy_id", policyID).Error("Failed to delete policy")
		return NewAccessControlError(ErrTypeStorageError, "failed to delete policy", err)
	}

	m.logger.WithField("policy_id", policyID).Info("Policy deleted successfully")
	return nil
}

func (m *rbacManager) ListPolicies(ctx context.Context) ([]*Policy, error) {
	policies, _, err := m.policyRepo.ListPolicies(ctx, PolicyFilters{})
	if err != nil {
		m.logger.WithError(err).Error("Failed to list policies")
		return nil, NewAccessControlError(ErrTypeStorageError, "failed to list policies", err)
	}

	return policies, nil
}

// Private helper methods

func (m *rbacManager) validateRole(role *Role) error {
	if role.Name == "" {
		return NewAccessControlError(ErrTypeInvalidRole, "role name cannot be empty", nil)
	}

	if role.Type == "" {
		role.Type = RoleTypeCustom
	}

	// Validate role type
	if role.Type != RoleTypeSystem && role.Type != RoleTypeTactical && role.Type != RoleTypeCustom {
		return NewAccessControlError(ErrTypeInvalidRole, "invalid role type", nil)
	}

	return nil
}

func (m *rbacManager) validatePermission(permission *Permission) error {
	if permission.Name == "" {
		return NewAccessControlError(ErrTypeInvalidPermission, "permission name cannot be empty", nil)
	}

	if permission.Resource == "" {
		return NewAccessControlError(ErrTypeInvalidPermission, "permission resource cannot be empty", nil)
	}

	if permission.Action == "" {
		return NewAccessControlError(ErrTypeInvalidPermission, "permission action cannot be empty", nil)
	}

	if permission.Effect == "" {
		permission.Effect = PermissionTypeAllow
	}

	if permission.Effect != PermissionTypeAllow && permission.Effect != PermissionTypeDeny {
		return NewAccessControlError(ErrTypeInvalidPermission, "invalid permission effect", nil)
	}

	return nil
}

func (m *rbacManager) validatePolicy(policy *Policy) error {
	if policy.Name == "" {
		return NewAccessControlError(ErrTypeInvalidPolicy, "policy name cannot be empty", nil)
	}

	if policy.Type == "" {
		policy.Type = PolicyTypeABAC
	}

	if policy.Effect == "" {
		policy.Effect = EffectTypeAllow
	}

	// Validate policy rules
	for i, rule := range policy.Rules {
		if rule.ID == "" {
			return NewAccessControlError(ErrTypeInvalidPolicy, fmt.Sprintf("rule %d missing ID", i), nil)
		}

		if rule.Condition == "" {
			return NewAccessControlError(ErrTypeInvalidPolicy, fmt.Sprintf("rule %s missing condition", rule.ID), nil)
		}

		// Validate condition syntax if evaluator is available
		if m.attributeEval != nil {
			if err := m.attributeEval.ValidateExpression(rule.Condition); err != nil {
				return NewAccessControlError(ErrTypeInvalidPolicy, 
					fmt.Sprintf("invalid condition in rule %s: %v", rule.ID, err), err)
			}
		}
	}

	return nil
}

func (m *rbacManager) checkRBACPermissions(ctx context.Context, request *AuthorizationRequest, roleBindings []*RoleBinding) (bool, []Evidence, error) {
	evidence := []Evidence{}

	for _, binding := range roleBindings {
		if !binding.IsActive {
			continue
		}

		// Check if binding has expired
		if binding.ExpiresAt != nil && binding.ExpiresAt.Before(time.Now()) {
			continue
		}

		// Get role details
		role, err := m.roleRepo.GetRole(ctx, binding.RoleID)
		if err != nil {
			continue // Skip roles that can't be loaded
		}

		if !role.IsActive {
			continue
		}

		// Check role permissions
		for _, permID := range role.Permissions {
			permission, err := m.permissionRepo.GetPermission(ctx, permID)
			if err != nil {
				continue // Skip permissions that can't be loaded
			}

			if !permission.IsActive {
				continue
			}

			// Check if permission matches request
			if permission.Resource == request.Resource && permission.Action == request.Action {
				evidence = append(evidence, Evidence{
					Type:      EvidenceTypePermission,
					Source:    permission.ID.String(),
					Value:     permission,
					Context:   map[string]interface{}{"role_id": role.ID.String()},
					Timestamp: time.Now(),
				})

				if permission.Effect == PermissionTypeAllow {
					return true, evidence, nil
				} else if permission.Effect == PermissionTypeDeny {
					return false, evidence, nil
				}
			}
		}

		// Add role evidence even if no matching permissions
		evidence = append(evidence, Evidence{
			Type:      EvidenceTypeRole,
			Source:    role.ID.String(),
			Value:     role,
			Timestamp: time.Now(),
		})
	}

	return false, evidence, nil
}

func (m *rbacManager) checkABACPolicies(ctx context.Context, request *AuthorizationRequest) (*PolicyDecision, error) {
	// Get active policies
	policies, err := m.policyRepo.GetActivePolicies(ctx)
	if err != nil {
		return nil, err
	}

	// Sort policies by priority (higher priority first)
	for i := 0; i < len(policies)-1; i++ {
		for j := i + 1; j < len(policies); j++ {
			if policies[i].Priority < policies[j].Priority {
				policies[i], policies[j] = policies[j], policies[i]
			}
		}
	}

	// Prepare attributes for evaluation
	attributes := make(map[string]interface{})
	
	// Add request attributes
	attributes["user_id"] = request.UserID.String()
	attributes["resource"] = request.Resource
	attributes["action"] = request.Action
	attributes["ip_address"] = request.IPAddress
	attributes["user_agent"] = request.UserAgent
	attributes["timestamp"] = request.Timestamp.Unix()
	
	// Add context attributes
	for k, v := range request.Context {
		attributes[k] = v
	}
	
	// Add user attributes
	for k, v := range request.Attributes {
		attributes[k] = v
	}

	// Evaluate policies
	for _, policy := range policies {
		decision, err := m.EvaluatePolicy(ctx, policy, attributes)
		if err != nil {
			m.logger.WithError(err).WithField("policy_id", policy.ID).Warn("Failed to evaluate policy")
			continue
		}

		if decision.Decision != DecisionTypeDeny {
			return decision, nil
		}
	}

	return nil, nil
}
