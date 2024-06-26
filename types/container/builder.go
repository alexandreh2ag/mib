package container

import "github.com/alexandreh2ag/mib/types"

type Builders map[string]BuilderImage

func (bs Builders) GetInstance(name string) BuilderImage {
	if instance, ok := bs[name]; ok {
		return instance
	}
	return nil
}

type BuilderImage interface {
	Type() string
	BuildImages(images types.Images, pushImages bool) error
	Build(image *types.Image, pushImages bool) error
	PushImages(images types.Images) error
	Push(tag string) error
}
