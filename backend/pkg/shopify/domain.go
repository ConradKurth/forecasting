package shopify

import "strings"

// NormalizeDomain ensures the domain always includes .myshopify.com
// This provides a consistent format for storage and usage
func NormalizeDomain(domain string) string {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return ""
	}
	
	// Remove any protocol if present
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	
	// Remove trailing slashes
	domain = strings.TrimSuffix(domain, "/")
	
	// Add .myshopify.com if not already present
	if !strings.HasSuffix(domain, ".myshopify.com") {
		domain = domain + ".myshopify.com"
	}
	
	return domain
}
