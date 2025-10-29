package utilities

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"math"
	mathrand "math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// UnifiedUtilityManager provides comprehensive utility management
type UnifiedUtilityManager struct {
	stringUtils     *StringUtils
	numberUtils     *NumberUtils
	dateUtils       *DateUtils
	cryptoUtils     *CryptoUtils
	validationUtils *ValidationUtils
	conversionUtils *ConversionUtils
	formatUtils     *FormatUtils
	fileUtils       *FileUtils
	networkUtils    *NetworkUtils
}

// NewUnifiedUtilityManager creates a new unified utility manager
func NewUnifiedUtilityManager() *UnifiedUtilityManager {
	return &UnifiedUtilityManager{
		stringUtils:     NewStringUtils(),
		numberUtils:     NewNumberUtils(),
		dateUtils:       NewDateUtils(),
		cryptoUtils:     NewCryptoUtils(),
		validationUtils: NewValidationUtils(),
		conversionUtils: NewConversionUtils(),
		formatUtils:     NewFormatUtils(),
		fileUtils:       NewFileUtils(),
		networkUtils:    NewNetworkUtils(),
	}
}

// GetStringUtils returns string utilities
func (uum *UnifiedUtilityManager) GetStringUtils() *StringUtils {
	return uum.stringUtils
}

// GetNumberUtils returns number utilities
func (uum *UnifiedUtilityManager) GetNumberUtils() *NumberUtils {
	return uum.numberUtils
}

// GetDateUtils returns date utilities
func (uum *UnifiedUtilityManager) GetDateUtils() *DateUtils {
	return uum.dateUtils
}

// GetCryptoUtils returns crypto utilities
func (uum *UnifiedUtilityManager) GetCryptoUtils() *CryptoUtils {
	return uum.cryptoUtils
}

// GetValidationUtils returns validation utilities
func (uum *UnifiedUtilityManager) GetValidationUtils() *ValidationUtils {
	return uum.validationUtils
}

// GetConversionUtils returns conversion utilities
func (uum *UnifiedUtilityManager) GetConversionUtils() *ConversionUtils {
	return uum.conversionUtils
}

// GetFormatUtils returns format utilities
func (uum *UnifiedUtilityManager) GetFormatUtils() *FormatUtils {
	return uum.formatUtils
}

// GetFileUtils returns file utilities
func (uum *UnifiedUtilityManager) GetFileUtils() *FileUtils {
	return uum.fileUtils
}

// GetNetworkUtils returns network utilities
func (uum *UnifiedUtilityManager) GetNetworkUtils() *NetworkUtils {
	return uum.networkUtils
}

// StringUtils provides string manipulation utilities
type StringUtils struct{}

// NewStringUtils creates a new string utils
func NewStringUtils() *StringUtils {
	return &StringUtils{}
}

// IsEmpty checks if string is empty
func (su *StringUtils) IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotEmpty checks if string is not empty
func (su *StringUtils) IsNotEmpty(s string) bool {
	return !su.IsEmpty(s)
}

// Truncate truncates string to specified length
func (su *StringUtils) Truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

// Capitalize capitalizes first letter
func (su *StringUtils) Capitalize(s string) string {
	if su.IsEmpty(s) {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// ToSnakeCase converts to snake_case
func (su *StringUtils) ToSnakeCase(s string) string {
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	return strings.ToLower(re.ReplaceAllString(s, "${1}_${2}"))
}

// ToCamelCase converts to camelCase
func (su *StringUtils) ToCamelCase(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) == 0 {
		return s
	}

	result := strings.ToLower(parts[0])
	for _, part := range parts[1:] {
		result += su.Capitalize(strings.ToLower(part))
	}
	return result
}

// ContainsAny checks if string contains any of the substrings
func (su *StringUtils) ContainsAny(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// RemoveWhitespace removes all whitespace
func (su *StringUtils) RemoveWhitespace(s string) string {
	return strings.ReplaceAll(s, " ", "")
}

// GenerateRandomString generates random string
func (su *StringUtils) GenerateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to simple random generation if crypto/rand fails
		for i := range bytes {
			bytes[i] = byte(mathrand.Intn(256))
		}
	}
	return hex.EncodeToString(bytes)[:length]
}

// NumberUtils provides number manipulation utilities
type NumberUtils struct{}

// NewNumberUtils creates a new number utils
func NewNumberUtils() *NumberUtils {
	return &NumberUtils{}
}

// IsEven checks if number is even
func (nu *NumberUtils) IsEven(n int) bool {
	return n%2 == 0
}

// IsOdd checks if number is odd
func (nu *NumberUtils) IsOdd(n int) bool {
	return n%2 != 0
}

// Clamp clamps number between min and max
func (nu *NumberUtils) Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Round rounds float to specified decimal places
func (nu *NumberUtils) Round(value float64, places int) float64 {
	multiplier := math.Pow(10, float64(places))
	return math.Round(value*multiplier) / multiplier
}

// FormatBytes formats bytes to human readable format
func (nu *NumberUtils) FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// DateUtils provides date manipulation utilities
type DateUtils struct{}

// NewDateUtils creates a new date utils
func NewDateUtils() *DateUtils {
	return &DateUtils{}
}

// IsToday checks if date is today
func (du *DateUtils) IsToday(date time.Time) bool {
	now := time.Now()
	return date.Year() == now.Year() && date.Month() == now.Month() && date.Day() == now.Day()
}

// IsYesterday checks if date is yesterday
func (du *DateUtils) IsYesterday(date time.Time) bool {
	yesterday := time.Now().AddDate(0, 0, -1)
	return date.Year() == yesterday.Year() && date.Month() == yesterday.Month() && date.Day() == yesterday.Day()
}

// FormatDuration formats duration to human readable format
func (du *DateUtils) FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	return fmt.Sprintf("%.1fd", d.Hours()/24)
}

// GetStartOfDay returns start of day
func (du *DateUtils) GetStartOfDay(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}

// GetEndOfDay returns end of day
func (du *DateUtils) GetEndOfDay(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())
}

// CryptoUtils provides cryptographic utilities
type CryptoUtils struct{}

// NewCryptoUtils creates a new crypto utils
func NewCryptoUtils() *CryptoUtils {
	return &CryptoUtils{}
}

// GenerateUUID generates a new UUID
func (cu *CryptoUtils) GenerateUUID() string {
	return uuid.New().String()
}

// GenerateRandomBytes generates random bytes
func (cu *CryptoUtils) GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	return bytes, err
}

// CalculateChecksum calculates CRC32 checksum
func (cu *CryptoUtils) CalculateChecksum(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}

// ValidationUtils provides validation utilities
type ValidationUtils struct{}

// NewValidationUtils creates a new validation utils
func NewValidationUtils() *ValidationUtils {
	return &ValidationUtils{}
}

// IsValidEmail validates email format
func (vu *ValidationUtils) IsValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// IsValidURL validates URL format
func (vu *ValidationUtils) IsValidURL(url string) bool {
	pattern := `^https?://[^\s/$.?#].[^\s]*$`
	matched, _ := regexp.MatchString(pattern, url)
	return matched
}

// IsValidUUID validates UUID format
func (vu *ValidationUtils) IsValidUUID(uuidStr string) bool {
	_, err := uuid.Parse(uuidStr)
	return err == nil
}

// IsValidPhoneNumber validates phone number format
func (vu *ValidationUtils) IsValidPhoneNumber(phone string) bool {
	pattern := `^\+?[1-9]\d{1,14}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

// ConversionUtils provides conversion utilities
type ConversionUtils struct{}

// NewConversionUtils creates a new conversion utils
func NewConversionUtils() *ConversionUtils {
	return &ConversionUtils{}
}

// StringToInt converts string to int
func (cu *ConversionUtils) StringToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// StringToFloat converts string to float64
func (cu *ConversionUtils) StringToFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// StringToBool converts string to bool
func (cu *ConversionUtils) StringToBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}

// IntToString converts int to string
func (cu *ConversionUtils) IntToString(i int) string {
	return strconv.Itoa(i)
}

// FloatToString converts float64 to string
func (cu *ConversionUtils) FloatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// BoolToString converts bool to string
func (cu *ConversionUtils) BoolToString(b bool) string {
	return strconv.FormatBool(b)
}

// FormatUtils provides formatting utilities
type FormatUtils struct{}

// NewFormatUtils creates a new format utils
func NewFormatUtils() *FormatUtils {
	return &FormatUtils{}
}

// FormatCurrency formats number as currency
func (fu *FormatUtils) FormatCurrency(amount float64, currency string) string {
	return fmt.Sprintf("%s %.2f", currency, amount)
}

// FormatPercentage formats number as percentage
func (fu *FormatUtils) FormatPercentage(value float64) string {
	return fmt.Sprintf("%.2f%%", value*100)
}

// FormatNumber formats number with thousand separators
func (fu *FormatUtils) FormatNumber(n int64) string {
	str := strconv.FormatInt(n, 10)
	if len(str) <= 3 {
		return str
	}

	var result strings.Builder
	for i, char := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(char)
	}
	return result.String()
}

// FileUtils provides file manipulation utilities
type FileUtils struct{}

// NewFileUtils creates a new file utils
func NewFileUtils() *FileUtils {
	return &FileUtils{}
}

// GetFileExtension gets file extension
func (fu *FileUtils) GetFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-1]
}

// GetFileName gets filename without extension
func (fu *FileUtils) GetFileName(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return filename
	}
	return strings.Join(parts[:len(parts)-1], ".")
}

// IsValidFileExtension checks if file extension is valid
func (fu *FileUtils) IsValidFileExtension(filename string, allowedExtensions []string) bool {
	extension := strings.ToLower(fu.GetFileExtension(filename))
	for _, allowed := range allowedExtensions {
		if extension == strings.ToLower(allowed) {
			return true
		}
	}
	return false
}

// NetworkUtils provides network utilities
type NetworkUtils struct{}

// NewNetworkUtils creates a new network utils
func NewNetworkUtils() *NetworkUtils {
	return &NetworkUtils{}
}

// IsValidIP validates IP address
func (nu *NetworkUtils) IsValidIP(ip string) bool {
	pattern := `^(\d{1,3}\.){3}\d{1,3}$`
	matched, _ := regexp.MatchString(pattern, ip)
	return matched
}

// IsValidPort validates port number
func (nu *NetworkUtils) IsValidPort(port int) bool {
	return port > 0 && port <= 65535
}

// ServiceManager provides service management functionality
type ServiceManager struct {
	services map[string]interface{}
	mu       sync.RWMutex
}

// NewServiceManager creates a new service manager
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		services: make(map[string]interface{}),
	}
}

// RegisterService registers a service
func (sm *ServiceManager) RegisterService(name string, service interface{}) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.services[name] = service
}

// GetService returns a service by name
func (sm *ServiceManager) GetService(name string) (interface{}, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	service, exists := sm.services[name]
	return service, exists
}

// GetAllServices returns all registered services
func (sm *ServiceManager) GetAllServices() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range sm.services {
		result[k] = v
	}
	return result
}

// UnregisterService removes a service
func (sm *ServiceManager) UnregisterService(name string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.services, name)
}

// ServiceRegistry provides service registry functionality
type ServiceRegistry struct {
	registry map[string]*ServiceInfo
	mu       sync.RWMutex
}

// ServiceInfo represents service information
type ServiceInfo struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Service      interface{}            `json:"service"`
	HealthCheck  func() error           `json:"-"`
	Metadata     map[string]interface{} `json:"metadata"`
	RegisteredAt time.Time              `json:"registered_at"`
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		registry: make(map[string]*ServiceInfo),
	}
}

// RegisterService registers a service with metadata
func (sr *ServiceRegistry) RegisterService(info *ServiceInfo) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	info.RegisteredAt = time.Now()
	sr.registry[info.Name] = info
}

// GetService returns service information
func (sr *ServiceRegistry) GetService(name string) (*ServiceInfo, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	info, exists := sr.registry[name]
	return info, exists
}

// GetAllServices returns all registered services
func (sr *ServiceRegistry) GetAllServices() map[string]*ServiceInfo {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	result := make(map[string]*ServiceInfo)
	for k, v := range sr.registry {
		result[k] = v
	}
	return result
}

// HealthCheck performs health check on all services
func (sr *ServiceRegistry) HealthCheck() map[string]error {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	results := make(map[string]error)
	for name, info := range sr.registry {
		if info.HealthCheck != nil {
			results[name] = info.HealthCheck()
		}
	}
	return results
}
