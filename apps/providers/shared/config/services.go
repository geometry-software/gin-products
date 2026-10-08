package config

import "strings"

// Service identifies a process in the backend service catalog.
type Service string

const (
	// Login owns user identity and directory endpoints.
	Login Service = "login"
	// Products owns the product catalog.
	Products Service = "products"
	// Shipping owns shipment state.
	Shipping Service = "shipping"
	// Invoices owns billing and invoice authorization.
	Invoices Service = "invoices"
	// Adapters owns MongoORM infrastructure endpoints.
	Adapters Service = "adapters"
	// Providers owns named infrastructure providers.
	Providers Service = "providers"
)

var defaultPorts = map[Service]string{
	Login: "3001", Products: "3002", Shipping: "3004",
	Invoices: "3005", Adapters: "3006", Providers: "3007",
}

// Port returns the configured port for a known service.
func Port(service string) string {
	fallback, found := defaultPorts[Service(service)]
	if !found {
		return ""
	}
	return Get(strings.ToUpper(service)+"_PORT", fallback)
}

// URL returns the configured base URL for a known service.
func URL(service string) string {
	return strings.TrimRight(Get(strings.ToUpper(service)+"_SERVICE_URL", "http://127.0.0.1:"+Port(service)), "/")
}
