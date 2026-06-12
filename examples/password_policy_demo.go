//go:build ignore
// +build ignore

// Package main demonstrates the password security policy implementation
// in GoTAK with comprehensive validation, scoring, and account lockout features.
package main

import (
	"fmt"

	"github.com/dfedick/gotak/internal/auth"
	"github.com/dfedick/gotak/pkg/logger"
)

func main() {
	// Initialize logger
	log := logger.NewDefault()
	log.Info().Msg("Starting password policy demonstration")

	// Create default military-grade password policy
	policy := auth.DefaultPasswordPolicy()
	validator := auth.NewPasswordValidator(policy, log)

	fmt.Println("=== GoTAK Password Security Policy Demo ===")

	// Display policy settings
	fmt.Printf("Password Policy Configuration:\n")
	fmt.Printf("  - Minimum Length: %d characters\n", policy.MinLength)
	fmt.Printf("  - Maximum Length: %d characters\n", policy.MaxLength)
	fmt.Printf("  - Require Uppercase: %v\n", policy.RequireUppercase)
	fmt.Printf("  - Require Lowercase: %v\n", policy.RequireLowercase)
	fmt.Printf("  - Require Numbers: %v\n", policy.RequireNumbers)
	fmt.Printf("  - Require Special Characters: %v\n", policy.RequireSpecialChars)
	fmt.Printf("  - Minimum Unique Characters: %d\n", policy.MinUniqueChars)
	fmt.Printf("  - Maximum Repeating Characters: %d\n", policy.MaxRepeatingChars)
	fmt.Printf("  - Forbid Sequential Characters: %v\n", policy.ForbidSequentialChars)
	fmt.Printf("  - Forbid Common Passwords: %v\n", policy.ForbidCommonPasswords)
	fmt.Printf("  - Allowed Special Characters: %s\n", policy.AllowedSpecialChars)
	fmt.Printf("  - Account Lockout After: %d attempts\n", policy.MaxFailedAttempts)
	fmt.Printf("  - Lockout Duration: %v\n", policy.LockoutDuration)
	fmt.Println()

	// Test various password examples
	testPasswords := []struct {
		password    string
		username    string
		description string
	}{
		{"password123", "johndoe", "Common weak password"},
		{"P@ssw0rd!", "johndoe", "Better but still contains sequential chars"},
		{"MyStr0ngP@ssw4rd!", "johndoe", "Strong password that meets all requirements"},
		{"VeryL0ngAndC0mpl3xP@ssw4rd2024!", "johndoe", "Very strong enterprise-grade password"},
		{"admin", "admin", "Extremely weak admin password"},
		{"johndoePassword97!", "johndoe", "Contains username (should fail)"},
		{"Sh0rt!", "johndoe", "Too short"},
		{"NoSpecialChars123", "johndoe", "Missing special characters"},
		{"nouppercase123!", "johndoe", "Missing uppercase letters"},
		{"NOLOWERCASE123!", "johndoe", "Missing lowercase letters"},
		{"NoNumbers!@#", "johndoe", "Missing numbers"},
		{"Rep3@ated!!!!", "johndoe", "Too many repeating characters"},
		{"SecureP@ssword456", "johndoe", "Contains sequential numbers"},
		{"UnicodeP@ssw4rd🔒97", "johndoe", "Password with unicode characters"},
	}

	fmt.Println("=== Password Validation Results ===")

	for i, test := range testPasswords {
		fmt.Printf("%d. Testing: \"%s\" (%s)\n", i+1, test.password, test.description)

		// Validate password
		err := validator.ValidatePassword(test.password, test.username)

		// Get strength metrics
		score := validator.GetPasswordStrengthScore(test.password)
		level := validator.GetPasswordStrengthLevel(test.password)

		// Generate detailed report
		report := validator.GeneratePasswordComplexityReport(test.password, test.username)

		fmt.Printf("   Validation: ")
		if err != nil {
			fmt.Printf("❌ FAILED - %s\n", err.Error())
		} else {
			fmt.Printf("✅ PASSED\n")
		}

		fmt.Printf("   Strength: %s (%d/100)\n", level, score)

		if len(report.Suggestions) > 0 {
			fmt.Printf("   Suggestions:\n")
			for _, suggestion := range report.Suggestions {
				fmt.Printf("     • %s\n", suggestion)
			}
		}

		// Show detailed checks for first few examples
		if i < 3 {
			fmt.Printf("   Detailed Checks:\n")
			for check, passed := range report.Checks {
				status := "❌"
				if passed {
					status = "✅"
				}
				fmt.Printf("     %s %s\n", status, check)
			}
		}

		fmt.Println()
	}

	// Demonstrate custom policy
	fmt.Println("=== Custom Lenient Policy Example ===")

	lenientPolicy := auth.PasswordPolicy{
		MinLength:                8,
		MaxLength:                64,
		RequireUppercase:         false,
		RequireLowercase:         true,
		RequireNumbers:           false,
		RequireSpecialChars:      false,
		MinUniqueChars:           4,
		ForbidCommonPasswords:    false,
		ForbidUsernameInPassword: false,
		MaxRepeatingChars:        5,
		ForbidSequentialChars:    false,
		AllowedSpecialChars:      "!@#$%^&*()",
		ForbiddenChars:           "",
		MaxFailedAttempts:        10,
		LockoutDuration:          policy.LockoutDuration,
	}

	lenientValidator := auth.NewPasswordValidator(lenientPolicy, log)

	simplePassword := "simplepass"
	fmt.Printf("Testing simple password \"%s\" with lenient policy:\n", simplePassword)

	err := lenientValidator.ValidatePassword(simplePassword, "testuser")
	if err != nil {
		fmt.Printf("❌ FAILED: %s\n", err.Error())
	} else {
		fmt.Printf("✅ PASSED\n")
	}

	score := lenientValidator.GetPasswordStrengthScore(simplePassword)
	level := lenientValidator.GetPasswordStrengthLevel(simplePassword)
	fmt.Printf("Strength: %s (%d/100)\n", level, score)

	fmt.Println("\n=== Security Best Practices ===")
	fmt.Println("1. Use a minimum of 12 characters for enterprise environments")
	fmt.Println("2. Include uppercase, lowercase, numbers, and special characters")
	fmt.Println("3. Avoid sequential patterns (abc, 123, qwerty)")
	fmt.Println("4. Don't include personal information like usernames")
	fmt.Println("5. Avoid common passwords and their variations")
	fmt.Println("6. Use unique passwords for each account")
	fmt.Println("7. Consider using a password manager for complex passwords")
	fmt.Println("8. Implement account lockout to prevent brute force attacks")
	fmt.Println("9. Enforce regular password changes (90 days recommended)")
	fmt.Println("10. Log all authentication attempts for security monitoring")

	fmt.Println("\n=== GoTAK Security Features ===")
	fmt.Println("✅ Comprehensive password complexity validation")
	fmt.Println("✅ Real-time password strength scoring (0-100)")
	fmt.Println("✅ Account lockout protection against brute force")
	fmt.Println("✅ Common password detection and blocking")
	fmt.Println("✅ Sequential pattern detection (abc, 123, qwerty)")
	fmt.Println("✅ Username inclusion prevention")
	fmt.Println("✅ Configurable policy settings per environment")
	fmt.Println("✅ Detailed validation feedback and suggestions")
	fmt.Println("✅ Military-grade default security policies")
	fmt.Println("✅ Comprehensive security audit logging")

	log.Info().Msg("Password policy demonstration completed successfully")
}
