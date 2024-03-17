package types

import "crypto/tls"

type UnitType uint32

const (
	UNIT_TYPE_HOST_MGR UnitType = 0
	UNIT_TYPE_HOST_SVC UnitType = 1
	UNIT_TYPE_HOST_APP UnitType = 2

	UNIT_TYPE_ADMVM     UnitType = 3
	UNIT_TYPE_ADMVM_MGR UnitType = 4
	UNIT_TYPE_ADMVM_SVC UnitType = 5
	UNIT_TYPE_ADMVM_APP UnitType = 6

	UNIT_TYPE_SYSVM     UnitType = 7
	UNIT_TYPE_SYSVM_MGR UnitType = 8
	UNIT_TYPE_SYSVM_SVC UnitType = 9
	UNIT_TYPE_SYSVM_APP UnitType = 10

	UNIT_TYPE_APPVM     UnitType = 11
	UNIT_TYPE_APPVM_MGR UnitType = 12
	UNIT_TYPE_APPVM_SVC UnitType = 13
	UNIT_TYPE_APPVM_APP UnitType = 14
)

type UnitStatus struct {
	Name        string
	Description string
	LoadState   string
	ActiveState string
	SubState    string
	Path        string
}

type TransportConfig struct {
	Address   string
	Port      string
	Protocol  string
	TlsConfig *tls.Config
}

type EndpointConfig struct {
	Name      string
	Transport TransportConfig
	Services  []string
}

type EndpointEntry struct {
	Name     string
	Protocol string
	Address  string
	Port     string
	WithTls  bool
}

type RegistryEntry struct {
	Name     string
	Parent   string
	Type     UnitType
	Endpoint EndpointEntry
	State    UnitStatus
	Watch    bool
}
