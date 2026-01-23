package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FormatBytes formats byte size to human-readable format
func FormatBytes(bytes int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	size := float64(bytes)
	unitIndex := 0

	for size >= 1024 && unitIndex < len(units)-1 {
		size /= 1024
		unitIndex++
	}

	if unitIndex == 0 {
		return fmt.Sprintf("%d %s", int64(size), units[unitIndex])
	}
	return fmt.Sprintf("%.2f %s", size, units[unitIndex])
}

// FormatDuration formats duration to human-readable format
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	if d < time.Hour {
		minutes := d.Minutes()
		return fmt.Sprintf("%.2fm", minutes)
	}
	hours := d.Hours()
	return fmt.Sprintf("%.2fh", hours)
}

// FormatTime formats time to ISO 8601 format
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// FormatTimeCustom formats time with custom layout
func FormatTimeCustom(t time.Time, layout string) string {
	return t.Format(layout)
}

// FormatDate formats time to date-only format (YYYY-MM-DD)
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatDateTime formats time to datetime format (YYYY-MM-DD HH:MM:SS)
func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatNumber formats number with comma separators
func FormatNumber(num int64) string {
	return strconv.FormatInt(num, 10)
}

// FormatPercentage formats float as percentage
func FormatPercentage(value float64) string {
	return fmt.Sprintf("%.2f%%", value)
}

// FormatCurrency formats float as currency
func FormatCurrency(amount float64, symbol string) string {
	return fmt.Sprintf("%s%.2f", symbol, amount)
}

// FormatJSON formats interface as indented JSON (for readability)
func FormatJSON(data interface{}) string {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", data)
	}
	return string(bytes)
}

// FormatPhoneNumber formats phone number
func FormatPhoneNumber(phone string) string {
	clean := RemoveSpecialChars(phone)
	if len(clean) < 10 {
		return phone
	}
	if len(clean) == 10 {
		return fmt.Sprintf("(%s) %s-%s", clean[0:3], clean[3:6], clean[6:10])
	}
	if len(clean) > 10 {
		return fmt.Sprintf("+%s (%s) %s-%s", clean[0:1], clean[1:4], clean[4:7], clean[7:10])
	}
	return phone
}

// FormatSSN formats social security number (XXX-XX-XXXX)
func FormatSSN(ssn string) string {
	clean := RemoveSpecialChars(ssn)
	if len(clean) != 9 {
		return ssn
	}
	return fmt.Sprintf("%s-%s-%s", clean[0:3], clean[3:5], clean[5:9])
}

// FormatCreditCard formats credit card number with mask
func FormatCreditCard(cardNumber string) string {
	clean := RemoveSpecialChars(cardNumber)
	if len(clean) < 4 {
		return clean
	}
	masked := ""
	for i := 0; i < len(clean)-4; i++ {
		masked += "*"
	}
	return masked + clean[len(clean)-4:]
}

// FormatEmail formats email by masking part of it
func FormatEmail(email string) string {
	if len(email) < 3 {
		return email
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	localPart := parts[0]
	domain := parts[1]

	if len(localPart) <= 2 {
		return fmt.Sprintf("%s***@%s", localPart[0:1], domain)
	}

	masked := string(localPart[0]) + "***" + string(localPart[len(localPart)-1])
	return fmt.Sprintf("%s@%s", masked, domain)
}

// FormatURL formats URL by removing query parameters
func FormatURL(urlStr string) string {
	parts := strings.Split(urlStr, "?")
	return parts[0]
}

// FormatPath formats file path consistently
func FormatPath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// FormatBoolean formats boolean as string
func FormatBoolean(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// FormatList formats slice as comma-separated string
func FormatList(items []string) string {
	return strings.Join(items, ", ")
}

// FormatKey formats string as object key (lowercase, underscores)
func FormatKey(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = RemoveSpecialChars(s)
	return s
}

// FormatTitle formats string as title (capitalize each word)
func FormatTitle(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		words[i] = CapitalizeString(word)
	}
	return strings.Join(words, " ")
}

// FormatCamelCase formats string as camelCase
func FormatCamelCase(s string) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}

	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		result += CapitalizeString(words[i])
	}
	return result
}

// FormatSnakeCase formats string as snake_case
func FormatSnakeCase(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "_")
	return s
}

// FormatKebabCase formats string as kebab-case
func FormatKebabCase(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

// FormatPascalCase formats string as PascalCase
func FormatPascalCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		words[i] = CapitalizeString(word)
	}
	return strings.Join(words, "")
}

// FormatMemoryUsage formats memory usage
func FormatMemoryUsage(bytes uint64) string {
	return FormatBytes(int64(bytes))
}

// FormatLoadAverage formats load average value
func FormatLoadAverage(load float64) string {
	return fmt.Sprintf("%.2f", load)
}

// FormatUptime formats uptime duration
func FormatUptime(seconds int64) string {
	d := time.Duration(seconds) * time.Second
	return FormatDuration(d)
}

// FormatLatency formats network latency
func FormatLatency(milliseconds float64) string {
	if milliseconds < 1000 {
		return fmt.Sprintf("%.2fms", milliseconds)
	}
	return fmt.Sprintf("%.2fs", milliseconds/1000)
}

// FormatBitrate formats bitrate
func FormatBitrate(bitsPerSecond float64) string {
	units := []string{"bps", "Kbps", "Mbps", "Gbps"}
	size := bitsPerSecond
	unitIndex := 0

	for size >= 1000 && unitIndex < len(units)-1 {
		size /= 1000
		unitIndex++
	}

	return fmt.Sprintf("%.2f %s", size, units[unitIndex])
}

// FormatCount formats large numbers with commas and labels
func FormatCount(count int64) string {
	if count < 1000 {
		return strconv.FormatInt(count, 10)
	}
	if count < 1000000 {
		return fmt.Sprintf("%.1fK", float64(count)/1000)
	}
	if count < 1000000000 {
		return fmt.Sprintf("%.1fM", float64(count)/1000000)
	}
	return fmt.Sprintf("%.1fB", float64(count)/1000000000)
}
