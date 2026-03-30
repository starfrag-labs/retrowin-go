package system

// SystemUser represents a user's membership in a system.
type SystemUser struct {
	id       int
	userID   string
	systemID string
	username string
}

// NewSystemUser creates a new SystemUser.
func NewSystemUser(
	id int,
	userID string,
	systemID string,
	username string,
) *SystemUser {
	return &SystemUser{
		id:       id,
		userID:   userID,
		systemID: systemID,
		username: username,
	}
}

// Getters
func (su *SystemUser) ID() int       { return su.id }
func (su *SystemUser) UserID() string { return su.userID }
func (su *SystemUser) SystemID() string { return su.systemID }
func (su *SystemUser) Username() string { return su.username }