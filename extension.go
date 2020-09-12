package main

// RevealInFileManager is a general function which calls platform-dependent implementations
// to reveal given path in its parent directory in OS file manager.
func RevealInFileManager(path string) {
	revealPath(path)
}
