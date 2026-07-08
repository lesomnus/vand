package config

type PackageConfig map[string]PackageConfigItem

type PackageConfigItem struct {
	// Hidden indicates whether the package is hidden from the pkg.go.dev index.
	// If true, the package will not be listed on pkg.go.dev, and its documentation will not be generated.
	Hidden bool

	// Target specifies the target platform or environment for the package.
	Target string
	// Branch specifies the branch of the repository to use for the package.
	Branch string

	// Root, if set, is the go-import prefix advertised for this package: the
	// path of the module that sits at the root of Target. Set it when this
	// package's go.mod lives in a subdirectory of Target; the subdirectory is
	// the package path with Root removed. It must be a path prefix of the
	// package path (an empty string denotes the domain root). Leave it nil for
	// packages whose go.mod is at the repository root.
	//
	// It is a pointer so that an explicit empty string (the domain root) is
	// distinguishable from being unset.
	Root *string
}
