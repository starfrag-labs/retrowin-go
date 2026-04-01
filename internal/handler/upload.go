package handler

import (
	"context"
	"net/url"
	"path"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	"github.com/starfrag-lab/retrowin-go/internal/application/storage"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode/content"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// InitiateUpload implements POST /fs/{systemId}/upload/initiate.
func (h *Handler) InitiateUpload(ctx context.Context, req *apiv1.InitiateUploadRequest, params apiv1.InitiateUploadParams) (apiv1.InitiateUploadRes, error) {
	if req.Path == "" || req.Path[0] != '/' {
		return nil, errors.BadRequest("path must be absolute (start with /)")
	}

	var contentType string
	if req.ContentType.Set {
		contentType = req.ContentType.Value
	}

	session, err := h.storageSvc.InitiateUpload(ctx, &storage.InitiateUploadCommand{
		SystemID:    params.SystemId,
		ContentType: contentType,
		Size:        req.Size,
	})
	if err != nil {
		return nil, h.domainError(err)
	}

	uploadURL, _ := url.Parse(session.UploadURL)

	return &apiv1.UploadSessionResponse{
		UploadSession: apiv1.UploadSession{
			ObjectId:  session.ObjectID,
			UploadUrl: *uploadURL,
			ExpiresAt: toTimestamp(session.ExpiresAt),
		},
	}, nil
}

// CompleteUpload implements POST /fs/{systemId}/upload/complete.
func (h *Handler) CompleteUpload(ctx context.Context, req *apiv1.CompleteUploadRequest, params apiv1.CompleteUploadParams) (apiv1.CompleteUploadRes, error) {
	// Validate and parse path
	if req.Path == "" || req.Path[0] != '/' {
		return nil, h.domainError(errors.BadRequest("path must be absolute (start with /)"))
	}

	dirPath := path.Dir(req.Path)
	fileName := path.Base(req.Path)

	mode := inode.ModeObject | inode.PermOwnerRW | inode.PermGroupRX | inode.PermOtherR
	if req.Mode.Set {
		mode = inode.ModeObject | int(req.Mode.Value)
	}

	result, err := h.storageSvc.CompleteUpload(ctx, &storage.CompleteUploadCommand{
		ObjectID: req.ObjectId,
		SystemID: params.SystemId,
		Mode:     mode,
	})
	if err != nil {
		return nil, h.domainError(err)
	}

	// Resolve parent directory and link the inode
	parentDir, err := h.fsSvc.ResolvePath(ctx, params.SystemId, dirPath)
	if err != nil {
		return nil, h.domainError(err)
	}

	// Check if entry already exists (overwrite case)
	existing, err := h.fsSvc.ResolvePath(ctx, params.SystemId, req.Path)
	if err == nil && existing != nil {
		// Unlink old entry from parent directory
		if err := h.fsSvc.Unlink(ctx, parentDir.ID(), fileName); err != nil {
			return nil, h.domainError(err)
		}
		// Delete old inode
		if err := h.fsSvc.Delete(ctx, existing.ID()); err != nil {
			return nil, h.domainError(err)
		}
	}

	entry := content.DirEntry{
		Name:     fileName,
		InodeID:  result.Inode.ID(),
		FileType: uint8(inode.ModeObject >> 12),
	}
	if err := h.fsSvc.Link(ctx, parentDir.ID(), entry); err != nil {
		return nil, h.domainError(err)
	}

	return &apiv1.InodeResponse{
		Inode: *h.toInode(result.Inode),
	}, nil
}

// GetDownloadUrl implements GET /fs/{systemId}/download.
func (h *Handler) GetDownloadUrl(ctx context.Context, params apiv1.GetDownloadUrlParams) (apiv1.GetDownloadUrlRes, error) {
	// First resolve the path to get inode ID
	in, err := h.fsSvc.ResolvePath(ctx, params.SystemId, params.Path)
	if err != nil {
		return nil, h.domainError(err)
	}

	downloadURL, err := h.storageSvc.GetDownloadURL(ctx, in.ID())
	if err != nil {
		return nil, h.domainError(err)
	}

	parsedURL, _ := url.Parse(downloadURL)

	return &apiv1.DownloadURLResponse{
		DownloadUrl: apiv1.DownloadURL{
			DownloadUrl: *parsedURL,
			ExpiresAt:   toTimestamp(in.Mtime()), // TODO: Use actual expiry
		},
	}, nil
}
