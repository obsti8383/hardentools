/*
    Hardentools
    Copyright (C) 2017  Claudio Guarnieri, Mariano Graziano, Florian Probst

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
    "fmt"
    "os/exec"
    "golang.org/x/sys/windows/registry"
)

func trigger_fileassoc(enable bool) {
    // HKEY_CURRENT_USER\SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\FileExts\.hta\OpenWithProgids
    key, _ := registry.OpenKey(registry.CURRENT_USER, "SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Explorer\\FileExts\\.hta\\OpenWithProgids", registry.ALL_ACCESS)

    if enable {
        events.AppendText("Enabling potentially malicious file associations\n")
        
        // Step 1: Reassociate system wide default 
        _, err := exec.Command("cmd.exe","/E:ON", "/C", "assoc .hta=htafile").Output()
        if err != nil {
            events.AppendText("error occured")
            events.AppendText(fmt.Sprintln("%s", err))
        }
        
        // Step 2 (Reassociate user defaults) is not necessary, since this is automatically done by Windows on first usage
    } else {
        events.AppendText("Disabling potentially malicious file associations\n")
        
        // Step 1: Remove association (system wide default)
        _, err := exec.Command("cmd.exe","/E:ON", "/C", "assoc .hta=").Output()
        if err != nil {
            events.AppendText("error occured")
            events.AppendText(fmt.Sprintln("%s", err))
        }
        
        // Step 2: Remove user association
        value_names, _ := key.ReadValueNames(100)   // just used "100" because there shouldn't be more entries (default is one entry)
        for _, value_name := range value_names {
            key.DeleteValue(value_name)
            //events.AppendText("Deleted value "+value_name+"\n")
        }
    }

    key.Close()
}
