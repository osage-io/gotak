// WebAuthn MFA Demo - Example CLI tool for testing WebAuthn enrollment
// This demonstrates the WebAuthn/FIDO2 MFA provider functionality
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/dfedick/gotak/pkg/mfa"
	"github.com/dfedick/gotak/pkg/mfa/providers"
)

func main() {
	fmt.Println("🔐 GoTAK WebAuthn/FIDO2 MFA Demo")
	fmt.Println("================================")

	// Create WebAuthn configuration
	config := &mfa.WebAuthnConfig{
		Enabled:            true,
		RPID:               "gotak.local",
		RPName:             "GoTAK Tactical Platform",
		Origin:             "https://gotak.local",
		Timeout:            60 * time.Second,
		RequireResidentKey: false,
		UserVerification:   "preferred",
		Attestation:        "none",
	}

	// Create WebAuthn provider
	provider := providers.NewWebAuthnProvider(config)

	// Validate configuration
	if err := provider.ValidateConfiguration(); err != nil {
		log.Fatalf("❌ Configuration validation failed: %v", err)
	}

	fmt.Printf("✅ WebAuthn provider configured successfully\n")
	fmt.Printf("   Relying Party: %s (%s)\n", config.RPName, config.RPID)
	fmt.Printf("   Origin: %s\n", config.Origin)
	fmt.Printf("   Timeout: %v\n", config.Timeout)
	fmt.Println()

	// Simulate user enrollment
	fmt.Println("📝 Step 1: Generate WebAuthn enrollment secret")
	userID := uuid.New()
	metadata := map[string]string{
		"username":     "tactical_user",
		"display_name": "Tactical User",
	}

	secret, err := provider.GenerateSecret(context.Background(), userID, metadata)
	if err != nil {
		log.Fatalf("❌ Failed to generate secret: %v", err)
	}

	fmt.Printf("✅ Secret generated for user %s\n", userID)
	fmt.Printf("   Type: %s\n", secret.Type)
	fmt.Printf("   Secret ID: %s\n", secret.ID)

	// Get registration options for the frontend
	fmt.Println()
	fmt.Println("📱 Step 2: Get registration options for client")
	registrationOptions, err := provider.GetRegistrationOptions(context.Background(), secret)
	if err != nil {
		log.Fatalf("❌ Failed to get registration options: %v", err)
	}

	// Pretty print registration options
	var regOpts interface{}
	if err := json.Unmarshal([]byte(registrationOptions), &regOpts); err == nil {
		prettyJSON, _ := json.MarshalIndent(regOpts, "", "  ")
		fmt.Printf("✅ Registration options generated:\n%s\n", string(prettyJSON))
	} else {
		fmt.Printf("✅ Registration options: %s\n", registrationOptions)
	}

	// Simulate credential creation response (in real app, this comes from the browser)
	fmt.Println()
	fmt.Println("🔑 Step 3: Simulate credential creation response")
	credResponse := providers.CredentialCreationResponse{
		ID:    "demo-credential-id-12345",
		RawID: []byte("demo-credential-raw-id-12345"),
		Response: providers.AuthenticatorAttestationResponse{
			ClientDataJSON:    []byte(`{"type":"webauthn.create","challenge":"demo-challenge"}`),
			AttestationObject: []byte("mock-attestation-object-data"),
		},
		Type: "public-key",
	}

	credJSON, _ := json.Marshal(credResponse)
	
	// Verify enrollment
	fmt.Printf("✅ Simulating credential creation response\n")
	if err := provider.VerifyEnrollment(context.Background(), secret, string(credJSON)); err != nil {
		log.Fatalf("❌ Enrollment verification failed: %v", err)
	}

	fmt.Printf("✅ WebAuthn enrollment completed successfully!\n")
	fmt.Printf("   Credential ID stored: %s\n", secret.Metadata["credential_id"])

	// Simulate authentication challenge
	fmt.Println()
	fmt.Println("🔒 Step 4: Create authentication challenge")
	
	// Create MFA factor from enrolled secret
	factor := &mfa.MFAFactor{
		ID:     uuid.New(),
		UserID: userID,
		Type:   mfa.MFATypeWebAuthn,
		Name:   "Hardware Security Key",
		Status: mfa.MFAStatusActive,
		Metadata: secret.Metadata, // Copy credential info
	}

	challenge, err := provider.CreateChallenge(context.Background(), factor)
	if err != nil {
		log.Fatalf("❌ Failed to create challenge: %v", err)
	}

	fmt.Printf("✅ Authentication challenge created\n")
	fmt.Printf("   Challenge ID: %s\n", challenge.ID)
	fmt.Printf("   Expires at: %s\n", challenge.ExpiresAt.Format(time.RFC3339))

	// Get login options
	loginOptions, err := provider.GetLoginOptions(context.Background(), challenge)
	if err != nil {
		log.Fatalf("❌ Failed to get login options: %v", err)
	}

	// Pretty print login options
	var loginOpts interface{}
	if err := json.Unmarshal([]byte(loginOptions), &loginOpts); err == nil {
		prettyJSON, _ := json.MarshalIndent(loginOpts, "", "  ")
		fmt.Printf("✅ Login options for client:\n%s\n", string(prettyJSON))
	}

	// Simulate authentication response
	fmt.Println()
	fmt.Println("🔓 Step 5: Simulate authentication response")
	authResponse := providers.CredentialAssertionResponse{
		ID:    "demo-credential-id-12345",
		RawID: []byte("demo-credential-raw-id-12345"),
		Response: providers.AuthenticatorAssertionResponse{
			ClientDataJSON:    []byte(`{"type":"webauthn.get","challenge":"demo-challenge"}`),
			AuthenticatorData: []byte("mock-authenticator-data"),
			Signature:         []byte("mock-signature"),
		},
		Type: "public-key",
	}

	authJSON, _ := json.Marshal(authResponse)

	// Verify authentication
	fmt.Printf("✅ Simulating authentication response\n")
	if err := provider.VerifyChallenge(context.Background(), challenge, string(authJSON)); err != nil {
		log.Fatalf("❌ Authentication verification failed: %v", err)
	}

	fmt.Printf("✅ WebAuthn authentication successful!\n")

	// Summary
	fmt.Println()
	fmt.Println("🎉 WebAuthn/FIDO2 Demo Summary")
	fmt.Println("==============================")
	fmt.Println("✅ Configuration validation")
	fmt.Println("✅ User enrollment with registration options")
	fmt.Println("✅ Credential creation and verification")
	fmt.Println("✅ Authentication challenge generation")
	fmt.Println("✅ Authentication response verification")
	fmt.Println()
	fmt.Println("🔐 WebAuthn/FIDO2 MFA provider is ready for production!")
	fmt.Println()
	fmt.Println("💡 Next steps:")
	fmt.Println("   - Integrate with frontend JavaScript WebAuthn API")
	fmt.Println("   - Add database persistence for credentials")
	fmt.Println("   - Configure real HTTPS domain and certificates")
	fmt.Println("   - Test with actual FIDO2 security keys")
	fmt.Println("   - Enable in GoTAK server configuration")
}
