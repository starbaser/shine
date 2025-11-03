package prism

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/starbased-co/shine/pkg/config"
)

// Manager handles prism discovery and lifecycle management
type Manager struct {
	searchPaths   []string          // Prism directories to search
	autoPath      bool              // Automatically add prism dirs to PATH
	discoveryMode DiscoveryMode     // How to discover prisms
	cache         map[string]string // name -> absolute path cache
}

// NewManager creates a new prism manager with the given configuration
func NewManager(searchPaths []string, autoPath bool) *Manager {
	return NewManagerWithMode(searchPaths, autoPath, DiscoveryAuto)
}

// NewManagerWithMode creates a manager with explicit discovery mode
func NewManagerWithMode(searchPaths []string, autoPath bool, mode DiscoveryMode) *Manager {
	pm := &Manager{
		searchPaths:   expandPaths(searchPaths),
		autoPath:      autoPath,
		discoveryMode: mode,
		cache:         make(map[string]string),
	}

	if autoPath {
		pm.augmentPATH()
	}

	return pm
}

// augmentPATH adds prism directories to the PATH environment variable
func (pm *Manager) augmentPATH() {
	currentPath := os.Getenv("PATH")
	pathParts := []string{}

	// Add prism directories to front of PATH
	for _, dir := range pm.searchPaths {
		if _, err := os.Stat(dir); err == nil {
			pathParts = append(pathParts, dir)
		}
	}

	// Append existing PATH
	if currentPath != "" {
		pathParts = append(pathParts, currentPath)
	}

	newPath := strings.Join(pathParts, string(os.PathListSeparator))
	os.Setenv("PATH", newPath)
}

// FindPrism discovers a prism binary by name
// Discovery priority:
// 1. Explicit binary path from config (if provided)
// 2. Cache lookup
// 3. Manifest-based discovery (if mode is manifest or auto)
// 4. PATH search (includes prism dirs if auto_path enabled)
// 5. Prism directory iteration (if auto_path disabled)
// 6. Shine executable directory (backward compatibility)
func (pm *Manager) FindPrism(name string, cfg *config.PrismConfig) (string, error) {
	// Level 1: Check for explicit binary path in config
	if cfg != nil && cfg.Binary != "" {
		// If binary field contains a path separator, treat as absolute/relative path
		if strings.Contains(cfg.Binary, string(os.PathSeparator)) {
			absPath, err := filepath.Abs(cfg.Binary)
			if err == nil && isExecutable(absPath) {
				pm.cache[name] = absPath
				return absPath, nil
			}
		}
		// Otherwise treat as binary name
		name = cfg.Binary
	}

	// Level 2: Check cache first
	if path, ok := pm.cache[name]; ok {
		if isExecutable(path) {
			return path, nil
		}
		delete(pm.cache, name) // Stale cache entry
	}

	// Level 3: Try manifest-based discovery
	if pm.discoveryMode == DiscoveryManifest || pm.discoveryMode == DiscoveryAuto {
		prismDir, manifest, err := FindManifestDir(pm.searchPaths, name)
		if err == nil {
			binaryPath := filepath.Join(prismDir, manifest.Prism.Binary)
			if isExecutable(binaryPath) {
				pm.cache[name] = binaryPath
				return binaryPath, nil
			}
		}
	}

	// If using strict manifest mode, stop here
	if pm.discoveryMode == DiscoveryManifest {
		return "", fmt.Errorf("prism %s not found via manifest discovery", name)
	}

	// Resolve binary name using naming convention
	binaryName := pm.resolveBinaryName(name)

	// Level 4: Check PATH (includes augmented prism dirs if auto_path enabled)
	if path, err := exec.LookPath(binaryName); err == nil {
		pm.cache[name] = path
		return path, nil
	}

	// Level 5: Search prism directories explicitly (if not using auto_path)
	if !pm.autoPath {
		for _, dir := range pm.searchPaths {
			prismPath := filepath.Join(dir, binaryName)
			if isExecutable(prismPath) {
				pm.cache[name] = prismPath
				return prismPath, nil
			}
		}
	}

	// Level 6: Check relative to shine executable (backward compatibility)
	exePath, err := os.Executable()
	if err == nil {
		binDir := filepath.Dir(exePath)
		prismPath := filepath.Join(binDir, binaryName)
		if isExecutable(prismPath) {
			pm.cache[name] = prismPath
			return prismPath, nil
		}
	}

	return "", fmt.Errorf("prism %s not found (binary: %s)", name, binaryName)
}

// resolveBinaryName converts prism name to binary name using naming convention
func (pm *Manager) resolveBinaryName(name string) string {
	// If already prefixed with "shine-", use as-is
	if strings.HasPrefix(name, "shine-") {
		return name
	}
	// Otherwise, apply naming convention
	return fmt.Sprintf("shine-%s", name)
}

// DiscoverAll finds all available prisms in search paths
func (pm *Manager) DiscoverAll() ([]string, error) {
	prisms := make(map[string]bool)

	for _, dir := range pm.searchPaths {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // Skip inaccessible directories
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()

			// Check naming convention: shine-*
			if strings.HasPrefix(name, "shine-") {
				fullPath := filepath.Join(dir, name)
				if isExecutable(fullPath) {
					// Extract prism name (remove "shine-" prefix)
					prismName := strings.TrimPrefix(name, "shine-")
					prisms[prismName] = true
				}
			}
		}
	}

	result := make([]string, 0, len(prisms))
	for name := range prisms {
		result = append(result, name)
	}

	return result, nil
}

// isExecutable checks if a file exists and is executable
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && info.Mode()&0111 != 0
}

// expandPaths expands ~ and environment variables in paths
func expandPaths(paths []string) []string {
	expanded := make([]string, 0, len(paths))

	for _, path := range paths {
		// Expand ~
		if strings.HasPrefix(path, "~/") {
			home, err := os.UserHomeDir()
			if err == nil {
				path = filepath.Join(home, path[2:])
			}
		}

		// Expand environment variables
		path = os.ExpandEnv(path)

		expanded = append(expanded, path)
	}

	return expanded
}
