package main

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/go-steputils/v2/cache"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/sethvargo/go-githubactions"
)

const (
	// State keys for passing data between restore and save steps
	stateCachePrimaryKey = "CACHE_KEY"
	stateCacheMatchedKey = "CACHE_RESULT"
)

func main() {
	action := githubactions.New()

	// Detect if we're in the post phase by checking for state variables
	// that were set during the main (restore) phase.
	isPostPhase := action.Getenv("STATE_"+stateCachePrimaryKey) != ""

	if isPostPhase {
		if err := runSave(action); err != nil {
			action.Fatalf("save failed: %v", err)
		}
	} else {
		if err := runRestore(action); err != nil {
			action.Fatalf("restore failed: %v", err)
		}
	}
}

func runRestore(action *githubactions.Action) error {
	logger := log.NewLogger()
	envRepo := env.NewRepository()
	cmdFactory := command.NewFactory(envRepo)

	// Get inputs from GitHub Actions
	primaryKey := action.GetInput("key")
	if primaryKey == "" {
		return fmt.Errorf("key is required")
	}

	restoreKeysInput := action.GetInput("restore-keys")
	failOnCacheMiss := parseBool(action.GetInput("fail-on-cache-miss"))
	lookupOnly := parseBool(action.GetInput("lookup-only"))
	verbose := parseBool(action.GetInput("verbose"))

	// Save the primary key in state for the save step
	action.SaveState(stateCachePrimaryKey, primaryKey)

	logger.EnableDebugLog(verbose)

	// Build the list of keys (primary key + restore keys)
	keys := []string{primaryKey}
	if restoreKeysInput != "" {
		restoreKeys := parseMultilineInput(restoreKeysInput)
		keys = append(keys, restoreKeys...)
	}

	action.Infof("Searching for cache with keys: %s", strings.Join(keys, ", "))

	if lookupOnly {
		// For lookup-only, we just check if cache exists without downloading
		// The Bitrise library doesn't have a direct lookup-only mode, so we'll do restore
		// and let it succeed/fail
		action.Infof("Lookup-only mode: checking if cache exists")
	}

	// Use Bitrise cache restorer
	restorer := cache.NewRestorer(envRepo, logger, cmdFactory, nil)
	err := restorer.Restore(cache.RestoreCacheInput{
		StepId:         "github-cache-restore",
		Verbose:        verbose,
		Keys:           keys,
		NumFullRetries: 3,
	})

	if err != nil {
		if failOnCacheMiss {
			return fmt.Errorf("failed to restore cache entry. Input key: %s, error: %w", primaryKey, err)
		}
		action.Infof("Cache not found for input keys: %s", strings.Join(keys, ", "))
		action.SetOutput("cache-hit", "false")
		return nil
	}

	// Check the BITRISE_CACHE_HIT env var set by the restorer
	cacheHit := envRepo.Get("BITRISE_CACHE_HIT")

	if cacheHit == "false" || cacheHit == "" {
		action.SetOutput("cache-hit", "false")
		action.Infof("Cache not found for input keys: %s", strings.Join(keys, ", "))
		return nil
	}

	// exact = exact match on primary key, partial = matched a restore key
	isExactMatch := cacheHit == "exact"
	action.SetOutput("cache-hit", fmt.Sprintf("%t", isExactMatch))

	if isExactMatch {
		// Save the matched key so save step knows to skip
		action.SaveState(stateCacheMatchedKey, primaryKey)
		action.Infof("Cache restored from key: %s (exact match)", primaryKey)
	} else {
		action.Infof("Cache restored from key (partial match)")
	}

	return nil
}

func runSave(action *githubactions.Action) error {
	logger := log.NewLogger()
	envRepo := env.NewRepository()
	pathChecker := pathutil.NewPathChecker()
	pathProvider := pathutil.NewPathProvider()
	pathModifier := pathutil.NewPathModifier()

	// Get the primary key from state (set during restore phase) or input
	primaryKey := action.Getenv("STATE_" + stateCachePrimaryKey)
	if primaryKey == "" {
		primaryKey = action.GetInput("key")
	}

	if primaryKey == "" {
		action.Warningf("Key is not specified.")
		return nil
	}

	// Check if we already had an exact cache hit (skip saving if so)
	matchedKey := action.Getenv("STATE_" + stateCacheMatchedKey)
	if matchedKey == primaryKey {
		action.Infof("Cache hit occurred on the primary key %s, not saving cache.", primaryKey)
		return nil
	}

	pathsInput := action.GetInput("path")
	if pathsInput == "" {
		return fmt.Errorf("path is required")
	}
	paths := parseMultilineInput(pathsInput)

	verbose := parseBool(action.GetInput("verbose"))
	logger.EnableDebugLog(verbose)

	action.Infof("Saving cache with key: %s", primaryKey)
	action.Infof("Paths: %s", strings.Join(paths, ", "))

	// Use Bitrise cache saver
	saver := cache.NewSaver(envRepo, logger, pathProvider, pathModifier, pathChecker, nil)
	err := saver.Save(cache.SaveCacheInput{
		StepId:           "github-cache-save",
		Verbose:          verbose,
		Key:              primaryKey,
		Paths:            paths,
		IsKeyUnique:      false,
		CompressionLevel: 3,
	})

	if err != nil {
		action.Warningf("Cache save failed: %v", err)
		return nil
	}

	action.Infof("Cache saved with key: %s", primaryKey)
	return nil
}

// parseMultilineInput parses a multiline input string into a slice of strings
func parseMultilineInput(input string) []string {
	if input == "" {
		return nil
	}

	lines := strings.Split(input, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Handle negation patterns (starting with !)
		trimmed = strings.TrimPrefix(trimmed, "! ")
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// parseBool parses a string to boolean
func parseBool(s string) bool {
	return strings.ToLower(s) == "true"
}
