package consts

var QmlFormatPackageName = map[string]string{
	"apt":    "qt6-tools-dev",
	"dnf":    "qmlformat",
	"pacman": "qt6-tools",
	"nix":    "nixpkgs.qt6.qttools",
	"binary": "qmlformat",
}

var PlasmoidPreviewPackageName = map[string]string{
	"apt":    "plasma-sdk",
	"dnf":    "plasma-sdk",
	"pacman": "plasma-sdk",
	"nix":    "nixpkgs.kdePackages.plasma-sdk",
	"binary": "plasmoidviewer",
}

var CurlPackageName = map[string]string{
	"apt":    "curl",
	"dnf":    "curl",
	"pacman": "curl",
	"nix":    "nixpkgs.curl",
	"binary": "curl",
}

var GettextPackageName = map[string]string{
	"apt":    "gettext",
	"dnf":    "gettext",
	"pacman": "gettext",
	"nix":    "nixpkgs.gettext",
	"binary": "xgettext",
}
