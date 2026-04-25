package fs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

func TestMkdir_RootPath(t *testing.T) {
	svc := NewService(nil, nil, nil, nil)

	_, err := svc.Mkdir(context.Background(), "sys", "/", 0755)
	assert.Error(t, err)
	assert.True(t, errors.IsBadRequest(err))
}
