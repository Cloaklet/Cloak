//+build !darwin,!linux

package extension

// TODO
func openPath(path string) {}

// TODO
func isFuseAvailable() bool {
	return false
}

// TODO
func locateLogDirectory() (string, error) {
	return "", fmt.Errorf("platform not supported")
}
