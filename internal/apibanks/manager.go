package apibanks

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// APIBank represents a collection of related data APIs
type APIBank struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace,omitempty"`
	Description string            `json:"description"`
	Owner       string            `json:"owner"` // Team or org
	Version     string            `json:"version"`
	APIs        []APIReference    `json:"apis"`
	Tags        []string          `json:"tags"`
	Labels      map[string]string `json:"labels,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// APIBankManager manages API banks
type APIBankManager struct {
	mu    sync.RWMutex
	banks map[string]*APIBank
}

// NewAPIBankManager creates a new API bank manager
func NewAPIBankManager() *APIBankManager {
	return &APIBankManager{
		banks: make(map[string]*APIBank),
	}
}

// CreateBank creates a new API bank
func (abm *APIBankManager) CreateBank(ctx context.Context, bank *APIBank) error {
	if bank.Name == "" {
		return fmt.Errorf("bank name is required")
	}

	abm.mu.Lock()
	defer abm.mu.Unlock()

	if _, exists := abm.banks[bank.Name]; exists {
		return fmt.Errorf("bank already exists: %s", bank.Name)
	}

	bank.CreatedAt = time.Now()
	bank.UpdatedAt = time.Now()
	abm.banks[bank.Name] = bank

	return nil
}

// GetBank retrieves a bank
func (abm *APIBankManager) GetBank(name string) *APIBank {
	abm.mu.RLock()
	defer abm.mu.RUnlock()
	return abm.banks[name]
}

// ListBanks lists all banks
func (abm *APIBankManager) ListBanks() []*APIBank {
	abm.mu.RLock()
	defer abm.mu.RUnlock()

	banks := make([]*APIBank, 0, len(abm.banks))
	for _, b := range abm.banks {
		banks = append(banks, b)
	}
	return banks
}

// AddAPIToBank adds an API to a bank
func (abm *APIBankManager) AddAPIToBank(ctx context.Context, bankName string, api APIReference) error {
	abm.mu.Lock()
	defer abm.mu.Unlock()

	bank, exists := abm.banks[bankName]
	if !exists {
		return fmt.Errorf("bank not found: %s", bankName)
	}

	// Check for duplicates
	for _, existing := range bank.APIs {
		if existing.Name == api.Name {
			return fmt.Errorf("API already in bank: %s", api.Name)
		}
	}

	bank.APIs = append(bank.APIs, api)
	bank.UpdatedAt = time.Now()
	return nil
}

// RemoveAPIFromBank removes an API from a bank
func (abm *APIBankManager) RemoveAPIFromBank(ctx context.Context, bankName, apiName string) error {
	abm.mu.Lock()
	defer abm.mu.Unlock()

	bank, exists := abm.banks[bankName]
	if !exists {
		return fmt.Errorf("bank not found: %s", bankName)
	}

	filtered := make([]APIReference, 0)
	found := false
	for _, api := range bank.APIs {
		if api.Name != apiName {
			filtered = append(filtered, api)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("API not found in bank: %s", apiName)
	}

	bank.APIs = filtered
	bank.UpdatedAt = time.Now()
	return nil
}

// DeleteBank removes a bank by name.
func (abm *APIBankManager) DeleteBank(name string) error {
	abm.mu.Lock()
	defer abm.mu.Unlock()
	if _, exists := abm.banks[name]; !exists {
		return ErrBankNotFound
	}
	delete(abm.banks, name)
	return nil
}

// GetAPIsByDataClass gets all APIs that expose a data class
func (abm *APIBankManager) GetAPIsByDataClass(dataClass string) []APIReference {
	abm.mu.RLock()
	defer abm.mu.RUnlock()

	apis := make([]APIReference, 0)
	for _, bank := range abm.banks {
		for _, api := range bank.APIs {
			for _, dc := range api.DataClasses {
				if dc == dataClass {
					apis = append(apis, api)
					break
				}
			}
		}
	}
	return apis
}

// GetAPIsByTag gets all banks with a specific tag
func (abm *APIBankManager) GetBanksByTag(tag string) []*APIBank {
	abm.mu.RLock()
	defer abm.mu.RUnlock()

	banks := make([]*APIBank, 0)
	for _, bank := range abm.banks {
		for _, t := range bank.Tags {
			if t == tag {
				banks = append(banks, bank)
				break
			}
		}
	}
	return banks
}

// GetAPIsByOwner gets all banks owned by a team
func (abm *APIBankManager) GetBanksByOwner(owner string) []*APIBank {
	abm.mu.RLock()
	defer abm.mu.RUnlock()

	banks := make([]*APIBank, 0)
	for _, bank := range abm.banks {
		if bank.Owner == owner {
			banks = append(banks, bank)
		}
	}
	return banks
}

// APIBankCatalog provides discovery capabilities
type APIBankCatalog struct {
	manager *APIBankManager
}

// NewAPIBankCatalog creates a new catalog
func NewAPIBankCatalog(manager *APIBankManager) *APIBankCatalog {
	return &APIBankCatalog{manager: manager}
}

// SearchByDataClass searches for APIs by data class
func (abc *APIBankCatalog) SearchByDataClass(dataClass string) []APIReference {
	return abc.manager.GetAPIsByDataClass(dataClass)
}

// SearchByOwner searches for banks by owner
func (abc *APIBankCatalog) SearchByOwner(owner string) []*APIBank {
	return abc.manager.GetBanksByOwner(owner)
}

// SearchByTag searches for banks by tag
func (abc *APIBankCatalog) SearchByTag(tag string) []*APIBank {
	return abc.manager.GetBanksByTag(tag)
}

// GetAllAPIs gets all APIs across all banks
func (abc *APIBankCatalog) GetAllAPIs() []APIReference {
	banks := abc.manager.ListBanks()
	apis := make([]APIReference, 0)
	for _, bank := range banks {
		apis = append(apis, bank.APIs...)
	}
	return apis
}

// GlobalAPIBankManager is the package-level API bank manager
var GlobalAPIBankManager = NewAPIBankManager()

// CreateBank creates a bank via global manager
func CreateBank(ctx context.Context, bank *APIBank) error {
	return GlobalAPIBankManager.CreateBank(ctx, bank)
}

// GetBank gets a bank via global manager
func GetBank(name string) *APIBank {
	return GlobalAPIBankManager.GetBank(name)
}

// ListBanks lists banks via global manager
func ListBanks() []*APIBank {
	return GlobalAPIBankManager.ListBanks()
}

// AddAPIToBank adds API to bank via global manager
func AddAPIToBank(ctx context.Context, bankName string, api APIReference) error {
	return GlobalAPIBankManager.AddAPIToBank(ctx, bankName, api)
}
