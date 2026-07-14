package bootstrap

// ConfigureRunUser is a no-op on Windows.
func ConfigureRunUser(Environment) error {
	return nil
}
