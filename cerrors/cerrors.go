package cerrors

import "github.com/pkg/errors"

var (
	EmptyVersion       = errors.New("need to set up version. [use SetVersion(version string) function with non-empty version value]")
	PluginHasNoVersion = errors.New("plugin is not support version")
)
