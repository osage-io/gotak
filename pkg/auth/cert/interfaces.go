package cert

import (
	"context"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CertificateValidator validates X.509 certificates for authentication
type CertificateValidator interface {
	// ValidateCertificate validates a client certificate and returns user identity
	ValidateCertificate(ctx context.Context, cert *x509.Certificate, chain []*x509.Certificate) (*CertIdentity, error)

	// ValidateCertificateChain validates the entire certificate chain
	ValidateCertificateChain(ctx context.Context, chain []*x509.Certificate) error

	// GetTrustedCAs returns the list of trusted Certificate Authorities
	GetTrustedCAs() []*x509.Certificate

	// AddTrustedCA adds a trusted Certificate Authority
	AddTrustedCA(ctx context.Context, ca *x509.Certificate) error

	// RemoveTrustedCA removes a trusted Certificate Authority
	RemoveTrustedCA(ctx context.Context, caID string) error

	// CheckRevocation checks if a certificate is revoked using CRL or OCSP
	CheckRevocation(ctx context.Context, cert *x509.Certificate) error
}

// CertIdentity represents a user identity extracted from a certificate
type CertIdentity struct {
	ID              uuid.UUID                      `json:"id" db:"id"`
	UserID          uuid.UUID                      `json:"user_id" db:"user_id"`
	CertType        CertificateType                `json:"cert_type" db:"cert_type"`
	SerialNumber    string                         `json:"serial_number" db:"serial_number"`
	Subject         CertificateSubject             `json:"subject" db:"subject"`
	Issuer          CertificateIssuer              `json:"issuer" db:"issuer"`
	Fingerprint     string                         `json:"fingerprint" db:"fingerprint"`
	NotBefore       time.Time                      `json:"not_before" db:"not_before"`
	NotAfter        time.Time                      `json:"not_after" db:"not_after"`
	KeyUsage        []string                       `json:"key_usage" db:"key_usage"`
	ExtKeyUsage     []string                       `json:"ext_key_usage" db:"ext_key_usage"`
	Attributes      map[string]string              `json:"attributes" db:"attributes"`
	Status          CertificateStatus              `json:"status" db:"status"`
	LastUsed        *time.Time                     `json:"last_used,omitempty" db:"last_used"`
	CreatedAt       time.Time                      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time                      `json:"updated_at" db:"updated_at"`
}

// CertificateSubject contains certificate subject information
type CertificateSubject struct {
	CommonName         string   `json:"common_name,omitempty"`
	Country            []string `json:"country,omitempty"`
	Organization       []string `json:"organization,omitempty"`
	OrganizationalUnit []string `json:"organizational_unit,omitempty"`
	Locality           []string `json:"locality,omitempty"`
	Province           []string `json:"province,omitempty"`
	StreetAddress      []string `json:"street_address,omitempty"`
	PostalCode         []string `json:"postal_code,omitempty"`
	SerialNumber       string   `json:"serial_number,omitempty"`
	ExtraNames         []string `json:"extra_names,omitempty"`
}

// CertificateIssuer contains certificate issuer information
type CertificateIssuer struct {
	CommonName         string   `json:"common_name,omitempty"`
	Country            []string `json:"country,omitempty"`
	Organization       []string `json:"organization,omitempty"`
	OrganizationalUnit []string `json:"organizational_unit,omitempty"`
	Locality           []string `json:"locality,omitempty"`
	Province           []string `json:"province,omitempty"`
	SerialNumber       string   `json:"serial_number,omitempty"`
}

// CertificateType represents the type of certificate
type CertificateType string

const (
	CertTypeCAC     CertificateType = "cac"     // Common Access Card
	CertTypePIV     CertificateType = "piv"     // Personal Identity Verification
	CertTypeX509    CertificateType = "x509"    // Standard X.509 certificate
	CertTypeCustom  CertificateType = "custom"  // Custom certificate format
)

// CertificateStatus represents the status of a certificate
type CertificateStatus string

const (
	CertStatusActive    CertificateStatus = "active"
	CertStatusRevoked   CertificateStatus = "revoked"
	CertStatusExpired   CertificateStatus = "expired"
	CertStatusSuspended CertificateStatus = "suspended"
)

// CertificateRepository manages certificate storage and retrieval
type CertificateRepository interface {
	// StoreCertIdentity stores a certificate identity
	StoreCertIdentity(ctx context.Context, identity *CertIdentity) error

	// GetCertIdentity retrieves a certificate identity by fingerprint
	GetCertIdentity(ctx context.Context, fingerprint string) (*CertIdentity, error)

	// GetCertIdentityBySerial retrieves a certificate identity by serial number
	GetCertIdentityBySerial(ctx context.Context, serialNumber string) (*CertIdentity, error)

	// GetCertIdentitiesByUserID retrieves all certificate identities for a user
	GetCertIdentitiesByUserID(ctx context.Context, userID uuid.UUID) ([]*CertIdentity, error)

	// UpdateCertIdentity updates a certificate identity
	UpdateCertIdentity(ctx context.Context, identity *CertIdentity) error

	// DeleteCertIdentity deletes a certificate identity
	DeleteCertIdentity(ctx context.Context, id uuid.UUID) error

	// ListCertIdentities lists all certificate identities with pagination
	ListCertIdentities(ctx context.Context, offset, limit int) ([]*CertIdentity, int, error)

	// StoreTrustedCA stores a trusted Certificate Authority
	StoreTrustedCA(ctx context.Context, ca *TrustedCA) error

	// GetTrustedCA retrieves a trusted Certificate Authority by ID
	GetTrustedCA(ctx context.Context, id uuid.UUID) (*TrustedCA, error)

	// ListTrustedCAs lists all trusted Certificate Authorities
	ListTrustedCAs(ctx context.Context) ([]*TrustedCA, error)

	// UpdateTrustedCA updates a trusted Certificate Authority
	UpdateTrustedCA(ctx context.Context, ca *TrustedCA) error

	// DeleteTrustedCA deletes a trusted Certificate Authority
	DeleteTrustedCA(ctx context.Context, id uuid.UUID) error
}

// TrustedCA represents a trusted Certificate Authority
type TrustedCA struct {
	ID           uuid.UUID   `json:"id" db:"id"`
	Name         string      `json:"name" db:"name"`
	Certificate  []byte      `json:"certificate" db:"certificate"`
	Fingerprint  string      `json:"fingerprint" db:"fingerprint"`
	Subject      string      `json:"subject" db:"subject"`
	Issuer       string      `json:"issuer" db:"issuer"`
	NotBefore    time.Time   `json:"not_before" db:"not_before"`
	NotAfter     time.Time   `json:"not_after" db:"not_after"`
	KeyUsage     []string    `json:"key_usage" db:"key_usage"`
	IsRoot       bool        `json:"is_root" db:"is_root"`
	CRLEndpoints []string    `json:"crl_endpoints" db:"crl_endpoints"`
	OCSPServers  []string    `json:"ocsp_servers" db:"ocsp_servers"`
	Status       CAStatus    `json:"status" db:"status"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
}

// CAStatus represents the status of a Certificate Authority
type CAStatus string

const (
	CAStatusActive     CAStatus = "active"
	CAStatusInactive   CAStatus = "inactive"
	CAStatusSuspended  CAStatus = "suspended"
	CAStatusRevoked    CAStatus = "revoked"
)

// RevocationChecker checks certificate revocation status
type RevocationChecker interface {
	// CheckCRL checks certificate against Certificate Revocation List
	CheckCRL(ctx context.Context, cert *x509.Certificate, crlURL string) error

	// CheckOCSP checks certificate using Online Certificate Status Protocol
	CheckOCSP(ctx context.Context, cert *x509.Certificate, issuer *x509.Certificate, ocspURL string) error

	// GetCRLList retrieves the current CRL for a CA
	GetCRLList(ctx context.Context, crlURL string) (*x509.RevocationList, error)
}

// CertificateExtractor extracts user information from certificates
type CertificateExtractor interface {
	// ExtractUserID extracts user identifier from certificate
	ExtractUserID(cert *x509.Certificate) (string, error)

	// ExtractAttributes extracts user attributes from certificate
	ExtractAttributes(cert *x509.Certificate) (map[string]string, error)

	// ExtractCertType determines the certificate type (CAC, PIV, etc.)
	ExtractCertType(cert *x509.Certificate) CertificateType

	// ValidateKeyUsage validates certificate key usage for authentication
	ValidateKeyUsage(cert *x509.Certificate) error
}

// CertificateAuthenticator orchestrates certificate authentication
type CertificateAuthenticator interface {
	// AuthenticateWithCertificate performs certificate-based authentication
	AuthenticateWithCertificate(ctx context.Context, cert *x509.Certificate, chain []*x509.Certificate) (*AuthResult, error)

	// EnrollCertificate enrolls a new certificate for a user
	EnrollCertificate(ctx context.Context, userID uuid.UUID, cert *x509.Certificate) (*CertIdentity, error)

	// RevokeCertificate revokes a certificate
	RevokeCertificate(ctx context.Context, certID uuid.UUID, reason string) error

	// GetCertificateInfo returns information about a certificate
	GetCertificateInfo(ctx context.Context, cert *x509.Certificate) (*CertIdentity, error)
}

// AuthResult represents the result of certificate authentication
type AuthResult struct {
	Success      bool              `json:"success"`
	UserID       uuid.UUID         `json:"user_id,omitempty"`
	CertIdentity *CertIdentity     `json:"cert_identity,omitempty"`
	Attributes   map[string]string `json:"attributes,omitempty"`
	Roles        []string          `json:"roles,omitempty"`
	Error        string            `json:"error,omitempty"`
	Timestamp    time.Time         `json:"timestamp"`
}

// CertificateError represents certificate-related errors
type CertificateError struct {
	Type    CertErrorType `json:"type"`
	Message string        `json:"message"`
	Cause   error         `json:"cause,omitempty"`
}

func (e *CertificateError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (cause: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func NewCertificateError(errType CertErrorType, message string, cause error) *CertificateError {
	return &CertificateError{
		Type:    errType,
		Message: message,
		Cause:   cause,
	}
}

// CertErrorType represents different types of certificate errors
type CertErrorType string

const (
	ErrTypeInvalidCertificate   CertErrorType = "invalid_certificate"
	ErrTypeUntrustedCA          CertErrorType = "untrusted_ca"
	ErrTypeCertificateExpired   CertErrorType = "certificate_expired"
	ErrTypeCertificateRevoked   CertErrorType = "certificate_revoked"
	ErrTypeInvalidKeyUsage      CertErrorType = "invalid_key_usage"
	ErrTypeExtractionFailed     CertErrorType = "extraction_failed"
	ErrTypeRevocationCheckFailed CertErrorType = "revocation_check_failed"
	ErrTypeChainValidationFailed CertErrorType = "chain_validation_failed"
)
