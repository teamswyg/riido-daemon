package main

const defaultReleaseAPIURL = "https://api.github.com/repos/teamswyg/riido-daemon/releases?per_page=1"

type githubRelease struct {
	TagName string               `json:"tag_name"`
	Draft   bool                 `json:"draft"`
	Assets  []githubReleaseAsset `json:"assets"`
}

type githubReleaseAsset struct {
	Name string `json:"name"`
}

func (r githubRelease) AssetNames() []string {
	out := make([]string, 0, len(r.Assets))
	for _, asset := range r.Assets {
		out = append(out, asset.Name)
	}
	return out
}
