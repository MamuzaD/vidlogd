package backup

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func (e Error) Error() string { return e.Message }

var (
	ErrNotInitialized = Error{
		Type:    "initialization",
		Message: "Repository not initialized",
		Code:    "NOT_INITIALIZED",
	}
	ErrRepositoryNotFound = Error{
		Type:    "repository",
		Message: "Repository not found or inaccessible",
		Code:    "REPO_NOT_FOUND",
	}
	ErrRepositoryNotConfigured = Error{
		Type:    "configuration",
		Message: "GitHub repository not configured",
		Code:    "REPO_NOT_CONFIGURED",
	}
	ErrNoWritePermission = Error{
		Type:    "permission",
		Message: "No write permission to repository",
		Code:    "NO_WRITE_PERMISSION",
	}
	ErrManualResolutionRequired = Error{
		Type:    "sync",
		Message: "Manual resolution required (local and remote have diverged)",
		Code:    "MANUAL_RESOLUTION_REQUIRED",
	}
	ErrConflictsDetected = Error{
		Type:    "conflict",
		Message: "Merge conflicts detected",
		Code:    "CONFLICTS_DETECTED",
	}
	ErrDirtyWorkingTree = Error{
		Type:    "sync",
		Message: "Working tree has uncommitted changes; resolve before syncing",
		Code:    "DIRTY_WORKING_TREE",
	}
)
