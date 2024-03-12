package container

import (
	mock_types_container "github.com/alexandreh2ag/mib/mock/types/container"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestBuilders_GetInstance_SuccessEmpty(t *testing.T) {
	b := Builders{}
	got := b.GetInstance("wrong")
	assert.Nil(t, got)
}

func TestBuilders_GetInstance_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	instance := mock_types_container.NewMockBuilderImage(ctrl)
	b := Builders{"foo": instance}
	got := b.GetInstance("foo")
	assert.Equal(t, instance, got)
}
