package utils

import (
	"regexp"
	"strings"
	"unicode"
)

// TrimSpaces removes leading and trailing whitespace from a string
func TrimSpaces(s string) string {
	return strings.TrimSpace(s)
}

// TrimAllSpaces removes all whitespace from a string
func TrimAllSpaces(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

// ToLowerCase converts string to lowercase
func ToLowerCase(s string) string {
	return strings.ToLower(s)
}

// ToUpperCase converts string to uppercase
func ToUpperCase(s string) string {
	return strings.ToUpper(s)
}

// ReverseString reverses a string
func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// ContainsSubstring checks if a string contains a substring
func ContainsSubstring(s, substring string) bool {
	return strings.Contains(s, substring)
}

// ReplaceString replaces all occurrences of old with new
func ReplaceString(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// SplitString splits a string by delimiter
func SplitString(s, delimiter string) []string {
	return strings.Split(s, delimiter)
}

// JoinStrings joins string slice with delimiter
func JoinStrings(strs []string, delimiter string) string {
	return strings.Join(strs, delimiter)
}

// IsEmpty checks if string is empty or only whitespace
func IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// IsNotEmpty checks if string has content
func IsNotEmpty(s string) bool {
	return len(strings.TrimSpace(s)) > 0
}

// StringLength returns the length of a string
func StringLength(s string) int {
	return len([]rune(s)) // Handles multi-byte characters correctly
}

// HasPrefix checks if string starts with prefix
func HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// HasSuffix checks if string ends with suffix
func HasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// RemovePrefix removes prefix from string if present
func RemovePrefix(s, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}

// RemoveSuffix removes suffix from string if present
func RemoveSuffix(s, suffix string) string {
	return strings.TrimSuffix(s, suffix)
}

// TruncateString truncates string to max length with ellipsis if needed
func TruncateString(s string, maxLength int) string {
	if StringLength(s) <= maxLength {
		return s
	}
	runes := []rune(s)
	if maxLength < 3 {
		return string(runes[:maxLength])
	}
	return string(runes[:maxLength-3]) + "..."
}

// RemoveSpecialChars removes all non-alphanumeric characters
func RemoveSpecialChars(s string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	return reg.ReplaceAllString(s, "")
}

// CapitalizeString capitalizes first letter of string
func CapitalizeString(s string) string {
	if IsEmpty(s) {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// CountOccurrences counts how many times substring appears in string
func CountOccurrences(s, substring string) int {
	return strings.Count(s, substring)
}

// ReplaceMultiple replaces multiple substrings
func ReplaceMultiple(s string, replacements map[string]string) string {
	replacer := strings.NewReplacer()
	pairs := []string{}
	for old, new := range replacements {
		pairs = append(pairs, old, new)
	}
	replacer = strings.NewReplacer(pairs...)
	return replacer.Replace(s)
}

// PadLeft pads string on the left with character
func PadLeft(s string, length int, padChar string) string {
	strLen := StringLength(s)
	if strLen >= length {
		return s
	}
	padding := strings.Repeat(padChar, length-strLen)
	return padding + s
}

// PadRight pads string on the right with character
func PadRight(s string, length int, padChar string) string {
	strLen := StringLength(s)
	if strLen >= length {
		return s
	}
	padding := strings.Repeat(padChar, length-strLen)
	return s + padding
}
