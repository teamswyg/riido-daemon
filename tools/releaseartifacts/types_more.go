package main

type installer struct {
	Command           string   `json:"command"`
	SupportedGOOS     []string `json:"supported_goos"`
	SupportedGOARCH   []string `json:"supported_goarch"`
	VersionEnv        string   `json:"version_env"`
	InstallDirEnv     string   `json:"install_dir_env"`
	DefaultInstallDir string   `json:"default_install_dir"`
}

type desktopMSIX struct {
	DownloadSource   string   `json:"download_source"`
	StorageRoot      string   `json:"storage_root"`
	RequiredEnv      []string `json:"required_env"`
	CDNLatestBaseURL string   `json:"cdn_latest_base_url"`
	CDNLatestPaths   []string `json:"cdn_latest_paths"`
}

type sourceCheck struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
}

type checkResult struct {
	Name   string `json:"name"`
	File   string `json:"file"`
	Pass   bool   `json:"pass"`
	Detail string `json:"detail,omitempty"`
}
