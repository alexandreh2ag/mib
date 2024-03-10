package types

type ImageName struct {
	Name string `yaml:"name"`
	Tag  string `yaml:"tag"`
}

func (imn ImageName) GetFullName() string {
	return imn.Name + ":" + imn.Tag
}

func (imn ImageName) GetName() string {
	return imn.Name
}

func (imn ImageName) GetTag() string {
	return imn.Tag
}

type Image struct {
	ImageName        `yaml:",inline"`
	Alias            []ImageName `yaml:"alias"`
	Path             string
	RelativeDir      string
	Parent           *Image
	Children         Images
	HasLocalParent   bool
	HasToBuild       bool
	HasParentToBuild bool
	EnvVariables     map[string]string `yaml:"envvars"`
	Packages         map[string]string `yaml:"packages"`
}

func (im Image) GetParents() Images {
	var parents Images
	if im.Parent != nil {
		parents = append(parents, im.Parent)
		parents = append(parents, im.Parent.GetParents()...)
	}

	return parents
}

func (im Image) GetAllEnvVar() map[string]string {
	envVars := make(map[string]string)
	var images Images
	images = append(images, &im)
	images = append(images, im.GetParents()...)
	for i := len(images) - 1; i >= 0; i-- {
		for k, v := range images[i].EnvVariables {
			envVars[k] = v
		}
	}

	return envVars
}

func (im Image) GetAllPackages() map[string]string {
	packagesVars := make(map[string]string)
	var images Images
	images = append(images, &im)
	images = append(images, im.GetParents()...)
	for i := len(images) - 1; i >= 0; i-- {
		for k, v := range images[i].Packages {
			packagesVars[k] = v
		}
	}

	return packagesVars
}

func (im Image) GetNames() []string {
	var names []string

	names = append(names, im.GetFullName())

	for _, alias := range im.Alias {
		names = append(names, alias.GetFullName())
	}
	return names
}
