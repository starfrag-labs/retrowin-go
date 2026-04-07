package dentry

import (
	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
)

type service struct {
	inodeSvc inode.InodeService
}

// NewService creates a new dentry service.
func NewService(inodeSvc inode.InodeService) DentryService {
	return &service{
		inodeSvc: inodeSvc,
	}
}
