package util

import (
	"math/rand"
	"regexp"
	"strings"
	"time"
)

// GenerateOrganizationCode generates a unique organization code from organization name
// Format: [5-char-from-name][3-random-digits]
// Example: "PT Beresin Tech" → "BERES123"
func GenerateOrganizationCode(organizationName string) string {
	// Clean the name: remove special chars, spaces, numbers
	cleanName := cleanOrganizationName(organizationName)

	// Take first 5 characters (pad with 'X' if shorter)
	namePrefix := extractNamePrefix(cleanName, 5)

	// Generate 3 random digits
	randomSuffix := generateRandomDigits(3)

	return namePrefix + randomSuffix
}

// cleanOrganizationName removes unwanted characters and normalizes the name
func cleanOrganizationName(name string) string {
	// Convert to uppercase
	name = strings.ToUpper(name)

	// Remove common business prefixes/suffixes
	businessTerms := []string{"PT", "CV", "UD", "TOKO", "STORE", "COMPANY", "CORP", "LTD", "INC"}
	for _, term := range businessTerms {
		name = strings.ReplaceAll(name, term, "")
	}

	// Keep only alphabetic characters
	reg := regexp.MustCompile(`[^A-Z]`)
	name = reg.ReplaceAllString(name, "")

	return name
}

// extractNamePrefix extracts the first N characters from cleaned name
func extractNamePrefix(cleanName string, length int) string {
	if len(cleanName) >= length {
		return cleanName[:length]
	}

	// Pad with 'X' if name is shorter than required length
	result := cleanName
	for len(result) < length {
		result += "X"
	}

	return result
}

// generateRandomDigits generates N random digits as string
func generateRandomDigits(length int) string {
	rand.Seed(time.Now().UnixNano())

	result := ""
	for i := 0; i < length; i++ {
		result += string(rune('0' + rand.Intn(10)))
	}

	return result
}

// ValidateOrganizationCode validates organization code format
func ValidateOrganizationCode(code string) bool {
	if len(code) != 8 {
		return false
	}

	// First 5 should be letters, last 3 should be digits
	namePattern := regexp.MustCompile(`^[A-Z]{5}[0-9]{3}$`)
	return namePattern.MatchString(code)
}

// GenerateUniqueOrganizationCode generates a unique code with collision checking
func GenerateUniqueOrganizationCode(organizationName string, existingCodes []string) string {
	baseCode := GenerateOrganizationCode(organizationName)

	// Check for collisions
	codeExists := func(code string) bool {
		for _, existing := range existingCodes {
			if existing == code {
				return true
			}
		}
		return false
	}

	// If no collision, return base code
	if !codeExists(baseCode) {
		return baseCode
	}

	// If collision exists, try different random suffixes
	namePrefix := baseCode[:5]
	for attempt := 0; attempt < 100; attempt++ { // Max 100 attempts
		randomSuffix := generateRandomDigits(3)
		newCode := namePrefix + randomSuffix

		if !codeExists(newCode) {
			return newCode
		}
	}

	// If still collision after 100 attempts, fallback to timestamp-based
	timestamp := time.Now().Unix() % 1000
	return namePrefix + padNumber(int(timestamp), 3)
}

// padNumber pads a number with leading zeros to specified length
func padNumber(num int, length int) string {
	str := string(rune('0' + num%1000))
	for len(str) < length {
		str = "0" + str
	}
	return str
}
