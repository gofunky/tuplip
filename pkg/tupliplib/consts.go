package tupliplib

const (
	// DockerRegistry is the Docker Hub registry URL.
	DockerRegistry = "https://registry-1.docker.io/"

	// VersionSeparator is the separator that separates the alias form the semantic version.
	VersionSeparator = ":"

	// WildcardDependency is the alias for a wildcard dependency to build a root tag vector
	// (i.e., semantic version without a prefix).
	WildcardDependency = "_"

	// VersionDot is the separator that separates the digits of a semantic version.
	VersionDot = "."

	// DockerTagSeparator is the separator that separates the sub tags in a Docker tag.
	DockerTagSeparator = "-"

	// DockerFromInstruction is the FROM instruction in Dockerfiles.
	DockerFromInstruction = "FROM"

	// DockerAs is the alias in FROM instructions in Dockerfiles.
	DockerAs = "as"

	// VectorSeparator is the default tag vector separator.
	VectorSeparator = " "

	// RepositorySeparator is the character that separates the organization from the name.
	RepositorySeparator = "/"
)
