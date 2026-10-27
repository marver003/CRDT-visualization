package web

import (
	"embed"
	"io/fs"
	"os/exec"
	"runtime"
)

//go:embed all:ui
var files embed.FS

func UI() fs.FS {
	sub, err := fs.Sub(files, "ui")
	if err != nil {
		panic(err)
	}

	return sub
}

func HasUI() bool {
	_, err := fs.Stat(UI(), "index.html")

	return err == nil
}

func OpenBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
