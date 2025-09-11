package cert

import (
	"crypto/x509"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// StandardExtractor implements CertificateExtractor for standard X.509 certificates
type StandardExtractor struct {
	config *ExtractorConfig
}

// ExtractorConfig configures the certificate extractor
type ExtractorConfig struct {
	// UserIDAttribute specifies which certificate attribute contains the user ID
	// Common values: "CN", "serialNumber", "emailAddress", "upn"
	UserIDAttribute string `yaml:"user_id_attribute" json:"user_id_attribute"`

	// UserIDPattern is a regex pattern to extract user ID from the attribute value
	UserIDPattern string `yaml:"user_id_pattern" json:"user_id_pattern"`

	// RequiredKeyUsages specifies required key usages for authentication
	RequiredKeyUsages []x509.KeyUsage `yaml:"required_key_usages" json:"required_key_usages"`

	// RequiredExtKeyUsages specifies required extended key usages
	RequiredExtKeyUsages []x509.ExtKeyUsage `yaml:"required_ext_key_usages" json:"required_ext_key_usages"`

	// AttributeMappings maps certificate attributes to user attributes
	AttributeMappings map[string]string `yaml:"attribute_mappings" json:"attribute_mappings"`

	// CAC/PIV specific configurations
	CACConfig *CACExtractorConfig `yaml:"cac_config" json:"cac_config"`
	PIVConfig *PIVExtractorConfig `yaml:"piv_config" json:"piv_config"`
}

// CACExtractorConfig configures CAC certificate extraction
type CACExtractorConfig struct {
	// FASC-N (Federal Agency Smart Credential Number) extraction
	ExtractFASCN bool `yaml:"extract_fascn" json:"extract_fascn"`

	// UUID extraction from CAC certificates
	ExtractUUID bool `yaml:"extract_uuid" json:"extract_uuid"`

	// CHUID (Card Holder Unique Identifier) extraction
	ExtractCHUID bool `yaml:"extract_chuid" json:"extract_chuid"`
}

// PIVExtractorConfig configures PIV certificate extraction
type PIVExtractorConfig struct {
	// PIV Card UUID extraction
	ExtractCardUUID bool `yaml:"extract_card_uuid" json:"extract_card_uuid"`

	// PIV Expiration Date extraction
	ExtractExpirationDate bool `yaml:"extract_expiration_date" json:"extract_expiration_date"`
}

// OID constants for CAC/PIV certificate extensions
var (
	// CAC OIDs
	OIDCACFascn = []int{2, 16, 840, 1, 101, 3, 6, 6}
	OIDCACUUID  = []int{2, 16, 840, 1, 101, 3, 6, 9, 1}
	OIDCACCuid  = []int{2, 16, 840, 1, 101, 3, 6, 9, 2}

	// PIV OIDs
	OIDPIVCardUUID = []int{2, 16, 840, 1, 101, 3, 6, 9, 16, 1, 1}
	OIDPIVExpiry   = []int{2, 16, 840, 1, 101, 3, 6, 9, 16, 1, 2}

	// Common subject alternative name OIDs
	OIDSubjectAltName         = []int{2, 5, 29, 17}
	OIDPrincipalName          = []int{1, 3, 6, 1, 4, 1, 311, 20, 2, 3} // Microsoft UPN
	OIDEmailAddress           = []int{1, 2, 840, 113549, 1, 9, 1}
)

// NewStandardExtractor creates a new standard certificate extractor
func NewStandardExtractor(config *ExtractorConfig) *StandardExtractor {
	if config == nil {
		config = &ExtractorConfig{
			UserIDAttribute: "CN",
			AttributeMappings: map[string]string{
				"CN": "common_name",
				"O":  "organization",
				"OU": "organizational_unit",
				"C":  "country",
				"L":  "locality",
				"ST": "state_province",
			},
		}
	}

	return &StandardExtractor{config: config}
}

// ExtractUserID extracts user identifier from certificate
func (e *StandardExtractor) ExtractUserID(cert *x509.Certificate) (string, error) {
	var userID string
	var err error

	switch strings.ToUpper(e.config.UserIDAttribute) {
	case "CN":
		userID = cert.Subject.CommonName
	case "SERIALNUMBER":
		userID = cert.Subject.SerialNumber
	case "EMAILADDRESS":
		userID, err = e.extractEmailFromSubject(cert)
	case "UPN":
		userID, err = e.extractUPNFromSAN(cert)
	default:
		// Try to extract from subject extra names
		userID, err = e.extractFromSubjectExtraNames(cert, e.config.UserIDAttribute)
	}

	if err != nil {
		return "", NewCertificateError(ErrTypeExtractionFailed, fmt.Sprintf("failed to extract user ID from attribute %s", e.config.UserIDAttribute), err)
	}

	if userID == "" {
		return "", NewCertificateError(ErrTypeExtractionFailed, fmt.Sprintf("user ID attribute %s is empty", e.config.UserIDAttribute), nil)
	}

	// Apply regex pattern if configured
	if e.config.UserIDPattern != "" {
		userID, err = e.applyPattern(userID, e.config.UserIDPattern)
		if err != nil {
			return "", NewCertificateError(ErrTypeExtractionFailed, "failed to apply user ID pattern", err)
		}
	}

	return userID, nil
}

// ExtractAttributes extracts user attributes from certificate
func (e *StandardExtractor) ExtractAttributes(cert *x509.Certificate) (map[string]string, error) {
	attributes := make(map[string]string)

	// Extract basic subject attributes
	e.extractSubjectAttributes(cert, attributes)

	// Extract issuer attributes
	e.extractIssuerAttributes(cert, attributes)

	// Extract certificate-specific attributes
	e.extractCertificateAttributes(cert, attributes)

	// Extract CAC-specific attributes
	if e.config.CACConfig != nil {
		e.extractCACAttributes(cert, attributes)
	}

	// Extract PIV-specific attributes
	if e.config.PIVConfig != nil {
		e.extractPIVAttributes(cert, attributes)
	}

	// Extract Subject Alternative Name attributes
	e.extractSANAttributes(cert, attributes)

	// Apply attribute mappings
	e.applyAttributeMappings(attributes)

	return attributes, nil
}

// ExtractCertType determines the certificate type (CAC, PIV, etc.)
func (e *StandardExtractor) ExtractCertType(cert *x509.Certificate) CertificateType {
	// Check for CAC-specific extensions
	if e.hasCACExtensions(cert) {
		return CertTypeCAC
	}

	// Check for PIV-specific extensions
	if e.hasPIVExtensions(cert) {
		return CertTypePIV
	}

	// Check issuer for government CAs
	issuer := cert.Issuer.String()
	if e.isGovernmentCA(issuer) {
		// Could be either CAC or PIV, default to X.509 for general government certs
		return CertTypeX509
	}

	// Default to standard X.509
	return CertTypeX509
}

// ValidateKeyUsage validates certificate key usage for authentication
func (e *StandardExtractor) ValidateKeyUsage(cert *x509.Certificate) error {
	// Check required key usages
	if len(e.config.RequiredKeyUsages) > 0 {
		for _, requiredUsage := range e.config.RequiredKeyUsages {
			if cert.KeyUsage&requiredUsage == 0 {
				return NewCertificateError(ErrTypeInvalidKeyUsage, fmt.Sprintf("certificate missing required key usage: %v", requiredUsage), nil)
			}
		}
	}

	// Check required extended key usages
	if len(e.config.RequiredExtKeyUsages) > 0 {
		certExtUsages := make(map[x509.ExtKeyUsage]bool)
		for _, usage := range cert.ExtKeyUsage {
			certExtUsages[usage] = true
		}

		for _, requiredUsage := range e.config.RequiredExtKeyUsages {
			if !certExtUsages[requiredUsage] {
				return NewCertificateError(ErrTypeInvalidKeyUsage, fmt.Sprintf("certificate missing required extended key usage: %v", requiredUsage), nil)
			}
		}
	}

	// Default validation for authentication certificates
	if cert.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		return NewCertificateError(ErrTypeInvalidKeyUsage, "certificate must have digital signature key usage for authentication", nil)
	}

	return nil
}

// Helper methods

func (e *StandardExtractor) extractEmailFromSubject(cert *x509.Certificate) (string, error) {
	// Look for email in subject extra names
	for _, name := range cert.Subject.Names {
		if name.Type.Equal(OIDEmailAddress) {
			if email, ok := name.Value.(string); ok {
				return email, nil
			}
		}
	}

	// Look in subject alternative names
	for _, email := range cert.EmailAddresses {
		return email, nil
	}

	return "", fmt.Errorf("no email address found in certificate")
}

func (e *StandardExtractor) extractUPNFromSAN(cert *x509.Certificate) (string, error) {
	// Look for UPN in subject alternative names
	for _, ext := range cert.Extensions {
		if ext.Id.Equal(OIDSubjectAltName) {
			// Parse SAN extension - this is a simplified version
			// In production, you'd need proper ASN.1 parsing
			return "", fmt.Errorf("UPN extraction from SAN not fully implemented")
		}
	}
	return "", fmt.Errorf("no UPN found in certificate")
}

func (e *StandardExtractor) extractFromSubjectExtraNames(cert *x509.Certificate, attribute string) (string, error) {
	for _, name := range cert.Subject.Names {
		if name.Type.String() == attribute {
			if value, ok := name.Value.(string); ok {
				return value, nil
			}
		}
	}
	return "", fmt.Errorf("attribute %s not found in subject", attribute)
}

func (e *StandardExtractor) applyPattern(input, pattern string) (string, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex pattern %s: %w", pattern, err)
	}

	matches := regex.FindStringSubmatch(input)
	if len(matches) < 2 {
		return input, nil // Return original if no capture groups
	}

	return matches[1], nil // Return first capture group
}

func (e *StandardExtractor) extractSubjectAttributes(cert *x509.Certificate, attributes map[string]string) {
	subject := cert.Subject

	if subject.CommonName != "" {
		attributes["subject_cn"] = subject.CommonName
	}
	if subject.SerialNumber != "" {
		attributes["subject_serial"] = subject.SerialNumber
	}
	if len(subject.Country) > 0 {
		attributes["subject_country"] = strings.Join(subject.Country, ",")
	}
	if len(subject.Organization) > 0 {
		attributes["subject_organization"] = strings.Join(subject.Organization, ",")
	}
	if len(subject.OrganizationalUnit) > 0 {
		attributes["subject_ou"] = strings.Join(subject.OrganizationalUnit, ",")
	}
	if len(subject.Locality) > 0 {
		attributes["subject_locality"] = strings.Join(subject.Locality, ",")
	}
	if len(subject.Province) > 0 {
		attributes["subject_province"] = strings.Join(subject.Province, ",")
	}
}

func (e *StandardExtractor) extractIssuerAttributes(cert *x509.Certificate, attributes map[string]string) {
	issuer := cert.Issuer

	if issuer.CommonName != "" {
		attributes["issuer_cn"] = issuer.CommonName
	}
	if len(issuer.Organization) > 0 {
		attributes["issuer_organization"] = strings.Join(issuer.Organization, ",")
	}
	if len(issuer.OrganizationalUnit) > 0 {
		attributes["issuer_ou"] = strings.Join(issuer.OrganizationalUnit, ",")
	}
	if len(issuer.Country) > 0 {
		attributes["issuer_country"] = strings.Join(issuer.Country, ",")
	}
}

func (e *StandardExtractor) extractCertificateAttributes(cert *x509.Certificate, attributes map[string]string) {
	attributes["serial_number"] = cert.SerialNumber.String()
	attributes["not_before"] = cert.NotBefore.Format("2006-01-02T15:04:05Z")
	attributes["not_after"] = cert.NotAfter.Format("2006-01-02T15:04:05Z")
	
	// Key usage
	keyUsages := e.keyUsageToStrings(cert.KeyUsage)
	if len(keyUsages) > 0 {
		attributes["key_usage"] = strings.Join(keyUsages, ",")
	}

	// Extended key usage
	extKeyUsages := e.extKeyUsageToStrings(cert.ExtKeyUsage)
	if len(extKeyUsages) > 0 {
		attributes["ext_key_usage"] = strings.Join(extKeyUsages, ",")
	}
}

func (e *StandardExtractor) extractCACAttributes(cert *x509.Certificate, attributes map[string]string) {
	if !e.config.CACConfig.ExtractFASCN && !e.config.CACConfig.ExtractUUID && !e.config.CACConfig.ExtractCHUID {
		return
	}

	for _, ext := range cert.Extensions {
		switch {
		case e.config.CACConfig.ExtractFASCN && ext.Id.Equal(OIDCACFascn):
			if fascn := e.parseASN1String(ext.Value); fascn != "" {
				attributes["cac_fascn"] = fascn
			}
		case e.config.CACConfig.ExtractUUID && ext.Id.Equal(OIDCACUUID):
			if uuid := e.parseASN1String(ext.Value); uuid != "" {
				attributes["cac_uuid"] = uuid
			}
		case e.config.CACConfig.ExtractCHUID && ext.Id.Equal(OIDCACCuid):
			if chuid := e.parseASN1String(ext.Value); chuid != "" {
				attributes["cac_chuid"] = chuid
			}
		}
	}
}

func (e *StandardExtractor) extractPIVAttributes(cert *x509.Certificate, attributes map[string]string) {
	if !e.config.PIVConfig.ExtractCardUUID && !e.config.PIVConfig.ExtractExpirationDate {
		return
	}

	for _, ext := range cert.Extensions {
		switch {
		case e.config.PIVConfig.ExtractCardUUID && ext.Id.Equal(OIDPIVCardUUID):
			if uuid := e.parseASN1String(ext.Value); uuid != "" {
				attributes["piv_card_uuid"] = uuid
			}
		case e.config.PIVConfig.ExtractExpirationDate && ext.Id.Equal(OIDPIVExpiry):
			if expiry := e.parseASN1String(ext.Value); expiry != "" {
				attributes["piv_expiry"] = expiry
			}
		}
	}
}

func (e *StandardExtractor) extractSANAttributes(cert *x509.Certificate, attributes map[string]string) {
	if len(cert.EmailAddresses) > 0 {
		attributes["email_addresses"] = strings.Join(cert.EmailAddresses, ",")
	}

	if len(cert.DNSNames) > 0 {
		attributes["dns_names"] = strings.Join(cert.DNSNames, ",")
	}

	if len(cert.IPAddresses) > 0 {
		var ips []string
		for _, ip := range cert.IPAddresses {
			ips = append(ips, ip.String())
		}
		attributes["ip_addresses"] = strings.Join(ips, ",")
	}

	if len(cert.URIs) > 0 {
		var uris []string
		for _, uri := range cert.URIs {
			uris = append(uris, uri.String())
		}
		attributes["uris"] = strings.Join(uris, ",")
	}
}

func (e *StandardExtractor) applyAttributeMappings(attributes map[string]string) {
	if e.config.AttributeMappings == nil {
		return
	}

	for sourceKey, targetKey := range e.config.AttributeMappings {
		if value, exists := attributes[sourceKey]; exists {
			attributes[targetKey] = value
			if sourceKey != targetKey {
				delete(attributes, sourceKey)
			}
		}
	}
}

func (e *StandardExtractor) hasCACExtensions(cert *x509.Certificate) bool {
	for _, ext := range cert.Extensions {
		if ext.Id.Equal(OIDCACFascn) || ext.Id.Equal(OIDCACUUID) || ext.Id.Equal(OIDCACCuid) {
			return true
		}
	}
	return false
}

func (e *StandardExtractor) hasPIVExtensions(cert *x509.Certificate) bool {
	for _, ext := range cert.Extensions {
		if ext.Id.Equal(OIDPIVCardUUID) || ext.Id.Equal(OIDPIVExpiry) {
			return true
		}
	}
	return false
}

func (e *StandardExtractor) isGovernmentCA(issuer string) bool {
	govPatterns := []string{
		"DoD",
		"Department of Defense",
		"US Government",
		"Federal",
		"GSA",
		"DISA",
		"CRL",
	}

	issuerUpper := strings.ToUpper(issuer)
	for _, pattern := range govPatterns {
		if strings.Contains(issuerUpper, strings.ToUpper(pattern)) {
			return true
		}
	}
	return false
}

func (e *StandardExtractor) keyUsageToStrings(usage x509.KeyUsage) []string {
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

func (e *StandardExtractor) extKeyUsageToStrings(usage []x509.ExtKeyUsage) []string {
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
			usages = append(usages, "unknown_" + strconv.Itoa(int(u)))
		}
	}

	return usages
}

func (e *StandardExtractor) parseASN1String(data []byte) string {
	// This is a simplified ASN.1 string parser
	// In production, you'd use proper ASN.1 parsing libraries
	if len(data) < 2 {
		return ""
	}

	// Simple check for ASN.1 OCTET STRING (0x04) or UTF8String (0x0C)
	if data[0] == 0x04 || data[0] == 0x0C {
		length := int(data[1])
		if len(data) >= 2+length {
			return string(data[2 : 2+length])
		}
	}

	return ""
}
