// Hardentools
// Copyright (C) 2017  Security Without Borders
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"os/exec"
)

// Disables SMB v1 via Powershell
func trigger_smbv1(harden bool) {
	if harden == false {
		events.AppendText("Restoring SMBv1 to default state\n")
		
		// Restore old state
		// TODO
		out, err := exec.Command("powershell.exe", "-Command", "Enable-WindowsOptionalFeature -Online -FeatureName smb1protocol -NoRestart").CombinedOutput()
		if err != nil {
			events.AppendText("error occured:\n")
			events.AppendText(string(out))
		}
	} else {
		events.AppendText("Hardening by SMBv1\n")
		
		// TODO: save old state
		// TODO
		
		// executes "powershell.exe -command "Disable-WindowsOptionalFeature -Online -FeatureName smb1protocol"
		/* C:\\Windows\\SysWOW64\\WindowsPowerShell\\v1.0\\*/
		out, err := exec.Command("powershell.exe", "-Command", "Disable-WindowsOptionalFeature -Online -FeatureName smb1protocol -NoRestart").CombinedOutput()
		if err != nil {
			events.AppendText("error occured:\n")
			events.AppendText(string(out))
		}
	}
}
