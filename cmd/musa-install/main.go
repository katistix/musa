package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	home, err := os.UserHomeDir()
	must(err)
	appRoot := filepath.Join(home, "Applications", "Musa.app")
	macOSDir := filepath.Join(appRoot, "Contents", "MacOS")
	resDir := filepath.Join(appRoot, "Contents", "Resources")
	must(os.MkdirAll(macOSDir, 0o755))
	must(os.MkdirAll(resDir, 0o755))

	must(run("go", "build", "-o", filepath.Join(macOSDir, "Musa"), "."))
	must(copyFile("assets/musa.icns", filepath.Join(resDir, "musa.icns")))
	must(copyFile("assets/icon.png", filepath.Join(resDir, "icon.png")))
	must(os.WriteFile(filepath.Join(appRoot, "Contents", "Info.plist"), []byte(infoPlist), 0o644))

	fmt.Println("Installed:", appRoot)
	fmt.Println("Open it from Finder or run: open ~/Applications/Musa.app")
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

const infoPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleDevelopmentRegion</key><string>en</string>
	<key>CFBundleExecutable</key><string>Musa</string>
	<key>CFBundleIconFile</key><string>musa.icns</string>
	<key>CFBundleIdentifier</key><string>works.earendil.musa</string>
	<key>CFBundleInfoDictionaryVersion</key><string>6.0</string>
	<key>CFBundleName</key><string>Musa</string>
	<key>CFBundlePackageType</key><string>APPL</string>
	<key>CFBundleShortVersionString</key><string>0.1.0</string>
	<key>CFBundleVersion</key><string>1</string>
	<key>LSApplicationCategoryType</key><string>public.app-category.music</string>
	<key>NSHighResolutionCapable</key><true/>
</dict>
</plist>
`
