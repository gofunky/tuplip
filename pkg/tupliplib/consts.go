package tupliplib

const (
	// Space depicts a simple space character.
	Space = " "

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

	// DockerArgInstruction is the ARG instruction in Dockerfiles.
	DockerArgInstruction = "ARG"

	// VersionArg is the argument name for the root version.
	VersionArg = "VERSION"

	// VersionInstruction is the Dockerfile's ARG VERSION.
	VersionInstruction = DockerArgInstruction + Space + VersionArg

	// RepositoryArg is the argument name for the root repository.
	RepositoryArg = "REPOSITORY"

	// RepositoryInstruction is the Dockerfile's ARG REPOSITORY.
	RepositoryInstruction = DockerArgInstruction + Space + RepositoryArg

	// DockerScratch is the empty Docker base image alias.
	DockerScratch = "scratch"

	// ScratchInstruction is a simple Docker FROM instruction using scratch only.
	ScratchInstruction = DockerFromInstruction + Space + DockerScratch

	// WildcardInstruction depicts a FROM instruction with a wildcard dependency.
	WildcardInstruction = DockerFromInstruction + Space + WildcardDependency + VersionSeparator

	// VersionChars are the characters that are used in a semantic version.
	VersionChars = "0123456789."

	// ArgEquation depicts the equals character in a Docker ARG instruction.
	ArgEquation = "="

	// DockerAs is the alias in FROM instructions in Dockerfiles.
	DockerAs = Space + "as" + Space

	// VectorSeparator is the default tag vector separator.
	VectorSeparator = " "

	// RepositorySeparator is the character that separates the organization from the name.
	RepositorySeparator = "/"

	// Dockerfile is the default name of a Dockerfile.
	Dockerfile = "Dockerfile"
)
