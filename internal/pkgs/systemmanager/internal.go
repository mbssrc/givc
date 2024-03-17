package systemmanager

import (
	"fmt"
	"givc/internal/pkgs/serviceclient"
	"givc/internal/pkgs/types"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func (svc *AdminService) getRemoteStatus(req *types.RegistryEntry) (*types.UnitStatus, error) {

	// Configure client endpoint
	ep := req.Endpoint
	if ep.Address == "" {
		parent := svc.Registry.GetEntryByName(req.Parent)
		if parent == nil {
			return nil, fmt.Errorf("entry has no parent or address")
		}
		ep = parent.Endpoint
	}

	cfgClient := &types.EndpointConfig{
		Name: req.Name,
		Transport: types.TransportConfig{
			Protocol:  ep.Protocol,
			Address:   ep.Address,
			Port:      ep.Port,
			TlsConfig: svc.TlsConfig,
		},
	}

	// Fetch status info
	resp, err := serviceclient.GetRemoteStatus(cfgClient, req.Name)
	if err != nil {
		log.Errorf("Couldn't retrieve unit status for %s: %v\n", req.Name, err)
		return nil, err
	}
	return resp, nil
}

func (svc *AdminService) startVM(name string) error {

	// Configure host endpoint
	entries := svc.Registry.GetEntryByType(types.UNIT_TYPE_HOST_MGR)
	if len(entries) < 1 {
		return fmt.Errorf("cannot start vm, required host manager not registered")
	}
	if len(entries) > 1 {
		return fmt.Errorf("more than one host manager registered: %v", entries) //oO
	}
	host := entries[0]
	clientCfg := &types.EndpointConfig{
		Name: host.Name,
		Transport: types.TransportConfig{
			Protocol:  host.Endpoint.Protocol,
			Address:   host.Endpoint.Address,
			Port:      host.Endpoint.Port,
			TlsConfig: svc.TlsConfig,
		},
	}

	// Check status and start service
	vmName := "microvm@" + name + "-vm.service"
	statusResponse, err := serviceclient.GetRemoteStatus(clientCfg, vmName)
	if err != nil {
		return fmt.Errorf("cannot retrieve vm status %s: %v", vmName, err)
	}
	if statusResponse.LoadState != "loaded" {
		return fmt.Errorf("vm %s not loaded", vmName)
	}
	if statusResponse.ActiveState != "active" {
		_, err := serviceclient.StartRemoteService(clientCfg, vmName)
		if err != nil {
			return err
		}
		time.Sleep(VM_STARTUP_TIME)
		statusResponse, err := serviceclient.GetRemoteStatus(clientCfg, vmName)
		if err != nil {
			return fmt.Errorf("cannot retrieve vm status for %s: %v", vmName, err)
		}
		if statusResponse.ActiveState != "active" {
			// @TODO this may throw an error currently if unit not yet started
			return fmt.Errorf("cannot start vm %s", vmName)
		}
	}
	return nil
}

func (svc *AdminService) sendSystemCommand(name string) error {

	// Get host entry
	entries := svc.Registry.GetEntryByType(types.UNIT_TYPE_HOST_MGR)
	if len(entries) < 1 {
		return fmt.Errorf("cannot start vm, required host manager not registered")
	}
	if len(entries) > 1 {
		return fmt.Errorf("more than one host manager registered: %v", entries) //oO
	}
	host := entries[0]

	// Configure host endpoint
	clientCfg := &types.EndpointConfig{
		Name: host.Name,
		Transport: types.TransportConfig{
			Protocol:  host.Endpoint.Protocol,
			Address:   host.Endpoint.Address,
			Port:      host.Endpoint.Port,
			TlsConfig: svc.TlsConfig,
		},
	}

	// Start unit
	_, err := serviceclient.StartRemoteService(clientCfg, name)
	if err != nil {
		return err
	}
	return nil
}

func (svc *AdminService) handleError(entry *types.RegistryEntry) {

	var err error
	switch entry.Type {
	case types.UNIT_TYPE_APPVM_APP:
		// Application handling
		err = svc.Registry.Deregister(entry)
		if err != nil {
			log.Warnf("cannot de-register service: %s", err)
		}
	case types.UNIT_TYPE_APPVM_MGR:
		fallthrough
	case types.UNIT_TYPE_SYSVM_MGR:
		// If agent is not found, re-start VM
		name := strings.Split(entry.Name, "-vm.service")[0]
		name = strings.Split(name, "givc-")[1]
		err = svc.startVM(name)
		if err != nil {
			log.Errorf("cannot start vm for %s", name)
		}
	}
}
