package cert

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// StandardValidator implements CertificateValidator for X.509 certificate validation
type StandardValidator struct {
	trustedCAs        map[string]*x509.Certificate
	repository        CertificateRepository
	extractor         CertificateExtractor
	revocationChecker RevocationChecker
	config            *ValidatorConfig
}

// ValidatorConfig configures the certificate validator
type ValidatorConfig struct {
	// RequireValidChain specifies whether the entire certificate chain must be valid
	RequireValidChain bool `yaml:"require_valid_chain" json:"require_valid_chain"`

	// CheckRevocation specifies whether to check certificate revocation status
	CheckRevocation bool `yaml:"check_revocation" json:"check_revocation"`

	// AllowExpiredCerts allows expired certificates (for testing/dev)
	AllowExpiredCerts bool `yaml:"allow_expired_certs" json:"allow_expired_certs"`

	// MaxChainLength specifies the maximum allowed certificate chain length
	MaxChainLength int `yaml:"max_chain_length" json:"max_chain_length"`

	// ClockSkewTolerance allows for time differences between client and server
	ClockSkewTolerance time.Duration `yaml:"clock_skew_tolerance" json:"clock_skew_tolerance"`

	// RevocationCacheTimeout specifies how long to cache revocation check results
	RevocationCacheTimeout time.Duration `yaml:"revocation_cache_timeout" json:"revocation_cache_timeout"`

	// RequiredKeyUsages specifies key usages that must be present
	RequiredKeyUsages []x509.KeyUsage `yaml:"required_key_usages" json:"required_key_usages"`

	// RequiredExtKeyUsages specifies extended key usages that must be present
	RequiredExtKeyUsages []x509.ExtKeyUsage `yaml:"required_ext_key_usages" json:"required_ext_key_usages"`

	// TrustedCAFingerprints is a list of trusted CA certificate fingerprints
	TrustedCAFingerprints []string `yaml:"trusted_ca_fingerprints" json:"trusted_ca_fingerprints"`

	// EnableStrictValidation enables additional validation checks
	EnableStrictValidation bool `yaml:"enable_strict_validation" json:"enable_strict_validation"`
}

// NewStandardValidator creates a new standard certificate validator
func NewStandardValidator(repository CertificateRepository, extractor CertificateExtractor, revocationChecker RevocationChecker, config *ValidatorConfig) *StandardValidator {
	if config == nil {
		config = &ValidatorConfig{
			RequireValidChain:       true,
			CheckRevocation:        true,
			AllowExpiredCerts:      false,
			MaxChainLength:         10,
			ClockSkewTolerance:     5 * time.Minute,
			RevocationCacheTimeout: 1 * time.Hour,
			EnableStrictValidation: true,
		}
	}

	validator := &StandardValidator{
		trustedCAs:        make(map[string]*x509.Certificate),
		repository:        repository,
		extractor:         extractor,
		revocationChecker: revocationChecker,
		config:           config,
	}

	// Load trusted CAs from repository
	validator.loadTrustedCAs(context.Background())

	return validator
}

// ValidateCertificate validates a client certificate and returns user identity
func (v *StandardValidator) ValidateCertificate(ctx context.Context, cert *x509.Certificate, chain []*x509.Certificate) (*CertIdentity, error) {
	// Basic certificate validation
	if err := v.validateCertificateBasics(cert); err != nil {
		return nil, err
	}

	// Validate certificate chain
	if v.config.RequireValidChain && len(chain) > 0 {
		if err := v.ValidateCertificateChain(ctx, chain); err != nil {
			return nil, NewCertificateError(ErrTypeChainValidationFailed, "certificate chain validation failed", err)
		}
	}

	// Check key usage
	if err := v.extractor.ValidateKeyUsage(cert); err != nil {
		return nil, err
	}

	// Check revocation status
	if v.config.CheckRevocation {
		if err := v.CheckRevocation(ctx, cert); err != nil {
			return nil, NewCertificateError(ErrTypeCertificateRevoked, "certificate revocation check failed", err)
		}
	}

	// Extract user attributes
	attributes, err := v.extractor.ExtractAttributes(cert)
	if err != nil {
		return nil, err
	}

	// Extract user ID and add to attributes
	userID, err := v.extractor.ExtractUserID(cert)
	if err != nil {
		return nil, err
	}
	attributes["user_id"] = userID

	// Determine certificate type
	certType := v.extractor.ExtractCertType(cert)

	// Create certificate identity
	identity := &CertIdentity{
		ID:           uuid.New(),
		CertType:     certType,
		SerialNumber: cert.SerialNumber.String(),
		Subject:      v.extractSubjectInfo(cert),
		Issuer:       v.extractIssuerInfo(cert),
		Fingerprint:  v.calculateFingerprint(cert),
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		KeyUsage:     v.keyUsageToStrings(cert.KeyUsage),
		ExtKeyUsage:  v.extKeyUsageToStrings(cert.ExtKeyUsage),
		Attributes:   attributes,
		Status:       CertStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Try to find existing certificate identity
	existingIdentity, err := v.repository.GetCertIdentity(ctx, identity.Fingerprint)
	if err == nil && existingIdentity != nil {
		// Update existing identity
		identity.ID = existingIdentity.ID
		identity.UserID = existingIdentity.UserID
		identity.CreatedAt = existingIdentity.CreatedAt
		identity.LastUsed = &time.Time{}
		*identity.LastUsed = time.Now()

		if err := v.repository.UpdateCertIdentity(ctx, identity); err != nil {
			// Log error but don't fail authentication
			fmt.Printf("Failed to update certificate identity: %v\n", err)
		}
	}

	return identity, nil
}

// ValidateCertificateChain validates the entire certificate chain
func (v *StandardValidator) ValidateCertificateChain(ctx context.Context, chain []*x509.Certificate) error {
	if len(chain) == 0 {
		return NewCertificateError(ErrTypeChainValidationFailed, "empty certificate chain", nil)
	}

	if len(chain) > v.config.MaxChainLength {
		return NewCertificateError(ErrTypeChainValidationFailed, fmt.Sprintf("certificate chain too long: %d > %d", len(chain), v.config.MaxChainLength), nil)
	}

	// Create certificate pool with trusted CAs
	roots := x509.NewCertPool()
	for _, ca := range v.trustedCAs {
		roots.AddCert(ca)
	}

	// Create intermediate certificate pool
	intermediates := x509.NewCertPool()
	if len(chain) > 1 {
		for _, cert := range chain[1:] {
			intermediates.AddCert(cert)
		}
	}

	// Verify certificate chain
	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	// Allow expired certificates if configured
	if v.config.AllowExpiredCerts {
		now := time.Now()
		opts.CurrentTime = now.Add(-v.config.ClockSkewTolerance)
	}

	_, err := chain[0].Verify(opts)
	if err != nil {
		return NewCertificateError(ErrTypeChainValidationFailed, "certificate chain verification failed", err)
	}

	return nil
}

// GetTrustedCAs returns the list of trusted Certificate Authorities
func (v *StandardValidator) GetTrustedCAs() []*x509.Certificate {
	cas := make([]*x509.Certificate, 0, len(v.trustedCAs))
	for _, ca := range v.trustedCAs {
		cas = append(cas, ca)
	}
	return cas
}

// AddTrustedCA adds a trusted Certificate Authority
func (v *StandardValidator) AddTrustedCA(ctx context.Context, ca *x509.Certificate) error {
	fingerprint := v.calculateFingerprint(ca)

	// Store in memory
	v.trustedCAs[fingerprint] = ca

	// Store in repository
	trustedCA := &TrustedCA{
		ID:           uuid.New(),
		Name:         ca.Subject.CommonName,
		Certificate:  ca.Raw,
		Fingerprint:  fingerprint,
		Subject:      ca.Subject.String(),
		Issuer:       ca.Issuer.String(),
		NotBefore:    ca.NotBefore,
		NotAfter:     ca.NotAfter,
		KeyUsage:     v.keyUsageToStrings(ca.KeyUsage),
		IsRoot:       ca.IsCA,
		CRLEndpoints: ca.CRLDistributionPoints,
		OCSPServers:  ca.OCSPServer,
		Status:       CAStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return v.repository.StoreTrustedCA(ctx, trustedCA)
}

// RemoveTrustedCA removes a trusted Certificate Authority
func (v *StandardValidator) RemoveTrustedCA(ctx context.Context, caID string) error {
	// Remove from memory
	delete(v.trustedCAs, caID)

	// Parse UUID if needed
	if uuid, err := uuid.Parse(caID); err == nil {
		return v.repository.DeleteTrustedCA(ctx, uuid)
	}

	// If not a UUID, assume it's a fingerprint and find the CA first
	cas, err := v.repository.ListTrustedCAs(ctx)
	if err != nil {
		return err
	}

	for _, ca := range cas {
		if ca.Fingerprint == caID {
			return v.repository.DeleteTrustedCA(ctx, ca.ID)
		}
	}

	return fmt.Errorf("trusted CA with ID %s not found", caID)
}

// CheckRevocation checks if a certificate is revoked using CRL or OCSP
func (v *StandardValidator) CheckRevocation(ctx context.Context, cert *x509.Certificate) error {
	if v.revocationChecker == nil {
		return nil // Skip revocation checking if no checker configured
	}

	// Check OCSP first (faster)
	if len(cert.OCSPServer) > 0 {
		// Find issuer certificate
		issuer := v.findIssuerCertificate(cert)
		if issuer != nil {
			for _, ocspURL := range cert.OCSPServer {
				if err := v.revocationChecker.CheckOCSP(ctx, cert, issuer, ocspURL); err != nil {
					// Continue to next OCSP server or try CRL
					continue
				}
				return nil // OCSP check passed
			}
		}
	}

	// Check CRL if OCSP is not available or failed
	if len(cert.CRLDistributionPoints) > 0 {
		for _, crlURL := range cert.CRLDistributionPoints {
			if err := v.revocationChecker.CheckCRL(ctx, cert, crlURL); err != nil {
				// Continue to next CRL
				continue
			}
			return nil // CRL check passed
		}
	}

	// If we're in strict mode, require successful revocation checking
	if v.config.EnableStrictValidation {
		return NewCertificateError(ErrTypeRevocationCheckFailed, "could not verify certificate revocation status", nil)
	}

	return nil
}

// Helper methods

func (v *StandardValidator) validateCertificateBasics(cert *x509.Certificate) error {
	now := time.Now()

	// Check expiration
	if !v.config.AllowExpiredCerts {
		if now.Before(cert.NotBefore.Add(-v.config.ClockSkewTolerance)) {
			return NewCertificateError(ErrTypeCertificateExpired, "certificate is not yet valid", nil)
		}
		if now.After(cert.NotAfter.Add(v.config.ClockSkewTolerance)) {
			return NewCertificateError(ErrTypeCertificateExpired, "certificate has expired", nil)
		}
	}

	// Check required key usages
	if len(v.config.RequiredKeyUsages) > 0 {
		for _, requiredUsage := range v.config.RequiredKeyUsages {
			if cert.KeyUsage&requiredUsage == 0 {
				return NewCertificateError(ErrTypeInvalidKeyUsage, fmt.Sprintf("certificate missing required key usage: %v", requiredUsage), nil)
			}
		}
	}

	// Check required extended key usages
	if len(v.config.RequiredExtKeyUsages) > 0 {
		certExtUsages := make(map[x509.ExtKeyUsage]bool)
		for _, usage := range cert.ExtKeyUsage {
			certExtUsages[usage] = true
		}

		for _, requiredUsage := range v.config.RequiredExtKeyUsages {
			if !certExtUsages[requiredUsage] {
				return NewCertificateError(ErrTypeInvalidKeyUsage, fmt.Sprintf("certificate missing required extended key usage: %v", requiredUsage), nil)
			}
		}
	}

	return nil
}

func (v *StandardValidator) loadTrustedCAs(ctx context.Context) error {
	if v.repository == nil {
		return nil
	}

	cas, err := v.repository.ListTrustedCAs(ctx)
	if err != nil {
		return err
	}

	for _, ca := range cas {
		if ca.Status == CAStatusActive {
			cert, err := x509.ParseCertificate(ca.Certificate)
			if err != nil {
				continue // Skip invalid certificates
			}
			v.trustedCAs[ca.Fingerprint] = cert
		}
	}

	return nil
}

func (v *StandardValidator) findIssuerCertificate(cert *x509.Certificate) *x509.Certificate {
	for _, ca := range v.trustedCAs {
		if cert.CheckSignatureFrom(ca) == nil {
			return ca
		}
	}
	return nil
}

func (v *StandardValidator) calculateFingerprint(cert *x509.Certificate) string {
	hash := sha256.Sum256(cert.Raw)
	return fmt.Sprintf("%x", hash)
}

func (v *StandardValidator) extractSubjectInfo(cert *x509.Certificate) CertificateSubject {
	return CertificateSubject{
		CommonName:         cert.Subject.CommonName,
		Country:            cert.Subject.Country,
		Organization:       cert.Subject.Organization,
		OrganizationalUnit: cert.Subject.OrganizationalUnit,
		Locality:           cert.Subject.Locality,
		Province:           cert.Subject.Province,
		SerialNumber:       cert.Subject.SerialNumber,
	}
}

func (v *StandardValidator) extractIssuerInfo(cert *x509.Certificate) CertificateIssuer {
	return CertificateIssuer{
		CommonName:         cert.Issuer.CommonName,
		Country:            cert.Issuer.Country,
		Organization:       cert.Issuer.Organization,
		OrganizationalUnit: cert.Issuer.OrganizationalUnit,
		Locality:           cert.Issuer.Locality,
		Province:           cert.Issuer.Province,
		SerialNumber:       cert.Issuer.SerialNumber,
	}
}

func (v *StandardValidator) keyUsageToStrings(usage x509.KeyUsage) []string {
	var usages []string

	if usage&x509.KeyUsageDigitalSignature != 0 {
		usages = append(usages, "digital_signature")
	}
	if usage&x509.KeyUsageContentCommitment != 0 {
		usages = append(usages, "content_commitment")
	}
	if usage&x509.KeyUsageKeyEncipherment != 0 {
		usages = append(usages, "key_encipherment")
	}
	if usage&x509.KeyUsageDataEncipherment != 0 {
		usages = append(usages, "data_encipherment")
	}
	if usage&x509.KeyUsageKeyAgreement != 0 {
		usages = append(usages, "key_agreement")
	}
	if usage&x509.KeyUsageCertSign != 0 {
		usages = append(usages, "cert_sign")
	}
	if usage&x509.KeyUsageCRLSign != 0 {
		usages = append(usages, "crl_sign")
	}
	if usage&x509.KeyUsageEncipherOnly != 0 {
		usages = append(usages, "encipher_only")
	}
	if usage&x509.KeyUsageDecipherOnly != 0 {
		usages = append(usages, "decipher_only")
	}

	return usages
}

func (v *StandardValidator) extKeyUsageToStrings(usage []x509.ExtKeyUsage) []string {
	var usages []string

	for _, u := range usage {
		switch u {
		case x509.ExtKeyUsageClientAuth:
			usages = append(usages, "client_auth")
		case x509.ExtKeyUsageServerAuth:
			usages = append(usages, "server_auth")
		case x509.ExtKeyUsageCodeSigning:
			usages = append(usages, "code_signing")
		case x509.ExtKeyUsageEmailProtection:
			usages = append(usages, "email_protection")
		case x509.ExtKeyUsageTimeStamping:
			usages = append(usages, "time_stamping")
		case x509.ExtKeyUsageOCSPSigning:
			usages = append(usages, "ocsp_signing")
		default:
			usages = append(usages, fmt.Sprintf("unknown_%d", int(u)))
		}
	}

	return usages
}
