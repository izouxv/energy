//----------------------------------------
//
// Copyright © yanghy. All Rights Reserved.
//
// Licensed under Apache License Version 2.0, January 2004
//
// https://www.apache.org/licenses/LICENSE-2.0
//
//----------------------------------------

package internal

import (
	"github.com/energye/energy/v2/cmd/internal/packager"
)

var CmdPackage = &Command{
	UsageLine: "package -p [path]",
	Short:     "Making an Installation Package",
	Long: `
	-p project path, default current path.
	.  Execute default command

Making an Installation Package.
	Windows: 
		Download: https://nsis.sourceforge.io/ 
		Install and configure to Path environment variable, use the makensis.exe command.
	Linux: 
		Creating deb installation packages using dpkg
	MacOS:
		Generate app package for energy
`,
}

func init() {
	CmdPackage.Run = runPackage
}

func runPackage(c *CommandConfig) error {
	if project, err := packager.NewProject(c.Package.Path); err != nil {
		return err
	} else {
		if err = packager.GeneraNSISInstaller(project); err != nil {
			return err
		}
	}
	return nil
}
