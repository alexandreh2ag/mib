package printer

import (
	"github.com/alexandreh2ag/mib/types"
	"github.com/fatih/color"
	"github.com/xlab/treeprint"
)

func DisplayImagesTree(images types.Images) string {
	tree := treeprint.New()
	displayImages(images, tree)
	return tree.String()
}

func displayImages(images types.Images, tree treeprint.Tree) {
	for _, image := range images {

		name := image.GetFullName()
		if image.HasToBuild {
			name = color.GreenString(name)
		}

		if len(image.Children) > 0 {
			node := tree.AddBranch(name)
			displayImages(image.Children, node)
		} else {
			tree.AddNode(name)
		}
	}
}
