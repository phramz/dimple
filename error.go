package dimple

import "errors"

var (
	// ErrCircularDependency is returned when a dependency cycle has been detected
	ErrCircularDependency = errors.New("circular dependency detected")
	// ErrUnknownService is returned if a requested service does not exist
	ErrUnknownService = errors.New("unknown service")
	// ErrServiceFactoryFailed is returned when the factory cannot instantiate the service
	ErrServiceFactoryFailed = errors.New("factory failed to instantiate service")
)
