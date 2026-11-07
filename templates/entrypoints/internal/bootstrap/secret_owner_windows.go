//go:build windows

package bootstrap

// Windows service accounts retain the ACL inherited from the image-created
// secrets_internal directory.
func ChownGeneratedSecrets(string) error { return nil }
