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

/**
Windows Defender Attack Surface Reduction (ASR)
needs Windows 10 >= 1709

More details here:
	https://docs.microsoft.com/en-us/windows/threat-protection/windows-defender-exploit-guard/attack-surface-reduction-exploit-guard
	https://docs.microsoft.com/en-us/windows/threat-protection/windows-defender-exploit-guard/enable-attack-surface-reduction
	https://docs.microsoft.com/en-us/windows/threat-protection/windows-defender-exploit-guard/evaluate-attack-surface-reduction

	One can use the "ExploitGuard ASR test tool" or https://demo.wd.microsoft.com/?ocid=cx-wddocs-testground from Microsoft (see third link) to verify that ASR is working
*/

import (
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/sys/windows/registry"
)

var ruleIdArray = []string{"BE9BA2D9-53EA-4CDC-84E5-9B1EEEE46550", //Block executable content from email client and webmail
	"D4F940AB-401B-4EFC-AADC-AD5F3C50688A", //Block Office applications from creating child processes
	"3B576869-A4EC-4529-8536-B80A7769E899", // Block Office applications from creating executable content
	"75668C1F-73B5-4CF0-BB93-3ECF5CB7CC84", // Block Office applications from injecting code into other processes
	"D3E037E1-3EB8-44C8-A917-57927947596D", // Block JavaScript or VBScript from launching downloaded executable content
	"5BEB7EFE-FD9A-4556-801D-275E5FFC04CC", // Block execution of potentially obfuscated scripts
	"92E97FA1-2EDF-4476-BDD6-9DD0B4DDDC7B"} // Block Win32 API calls from Office macro
var ruleIDEnumeration = strings.Join(ruleIdArray, ",")

var actionsArray = []string{"Enabled", "Enabled", "Enabled", "Enabled", "Enabled", "Enabled", "Enabled"}
var actionsEnumeration = strings.Join(actionsArray, ",")

// data type for a RegEx Path / Single Value DWORD combination
type WindowsASRStruct struct {
	shortName   string
	longName    string
	description string
}

var WindowsASR = &WindowsASRStruct{
	shortName:   "WindowsASR",
	longName:    "Windows ASR (needs Win 10/1709)",
	description: "Windows Attack Surface Reduction (ASR) (needs Win 10/1709)",
}

//// HardenInterface methods

func (asr WindowsASRStruct) Harden(harden bool) error {
	if harden {
		// harden (but only if we have at least Windows 10 - 1709)
		if checkWindowsVersion() {
			psString := fmt.Sprintf("Set-MpPreference -AttackSurfaceReductionRules_Ids %s -AttackSurfaceReductionRules_Actions %s", ruleIDEnumeration, actionsEnumeration)
			_, err := executeCommand("PowerShell.exe", "-Command", psString)
			if err != nil {
				return HardenError{"!! Executing powershell cmdlet Set-MpPreference failed.\n"}
			}
		}
	} else {
		// restore (but only if we have at least Windows 10 - 1709)
		if checkWindowsVersion() {
			// This is how we switch off ASR again:
			//   Remove-MpPreference -AttackSurfaceReductionRules_Ids <ID1>, <ID2>, ...
			psString := fmt.Sprintf("Remove-MpPreference -AttackSurfaceReductionRules_Ids %s", ruleIDEnumeration)
			_, err := executeCommand("PowerShell.exe", "-Command", psString)
			if err != nil {
				return HardenError{"!! Executing powershell cmdlet Remove-MpPreference failed.\n"}
			}
		}
	}

	return nil
}

func (asr WindowsASRStruct) IsHardened() bool {
	var hardened = false

	if checkWindowsVersion() {
		// call "$prefs = Get-MpPreference; $prefs.AttackSurfaceReductionRules_Ids"
		psString := fmt.Sprintf("$prefs = Get-MpPreference; $prefs.AttackSurfaceReductionRules_Ids")
		ruleIDsOut, err := executeCommand("PowerShell.exe", "-Command", psString)
		if err != nil {
			return false // in case command does not work we assume we are not hardened
		}

		// call "$prefs = Get-MpPreference; $prefs.AttackSurfaceReductionRules_Actions"
		psString = fmt.Sprintf("$prefs = Get-MpPreference; $prefs.AttackSurfaceReductionRules_Actions")
		ruleActionsOut, err := executeCommand("PowerShell.exe", "-Command", psString)
		if err != nil {
			return false // in case command does not work we assume we are not hardened
		}

		//// verify if all relevant ruleIDs are there
		// split / remove line feeds and carriage return
		currentRuleIDs := strings.Split(ruleIDsOut, "\r\n")
		currentRuleActions := strings.Split(ruleActionsOut, "\r\n")

		// just some debug
		/*for i, ruleIDdebug := range currentRuleIDs {
			if len(ruleIDdebug) > 0 {
				fmt.Printf("ruleID %d = %s with action = %s\n", i, ruleIDdebug, currentRuleActions[i])
			}
		}*/

		// compare to hardened state
		for i, ruleIdHardened := range ruleIdArray {
			// check if rule exists by iterating over all ruleIDs
			var existsAndEqual = false

			for j, currentRuleID := range currentRuleIDs {
				if ruleIdHardened == currentRuleID {
					// verify if setting is the same (TODO: currently works only with "Enabled")
					if currentRuleActions[j] == "1" && actionsArray[i] == "Enabled" {
						// everything is fine
						existsAndEqual = true
					} else {
						// break here
						return false
					}
				}
			}

			if existsAndEqual == false {
				// break here
				return false
			}
		}

		return true // seems all relevant hardening is in place

		/* Unmodifed State in Windows 10 / 1709:
			   PS> Get-MpPreference
			    AttackSurfaceReductionOnlyExclusions          :
			    AttackSurfaceReductionRules_Actions           :
			    AttackSurfaceReductionRules_Ids               :

		      Modified State:
			    PS > $prefs = Get-MpPreference
				PS > $prefs.AttackSurfaceReductionOnlyExclusions
				PS > $prefs.AttackSurfaceReductionRules_Actions
					1
					1
					1
					1
					1
					1
					1
				PS > $prefs.AttackSurfaceReductionRules_Ids
					3B576869-A4EC-4529-8536-B80A7769E899
					5BEB7EFE-FD9A-4556-801D-275E5FFC04CC
					75668C1F-73B5-4CF0-BB93-3ECF5CB7CC84
					92E97FA1-2EDF-4476-BDD6-9DD0B4DDDC7B
					BE9BA2D9-53EA-4CDC-84E5-9B1EEEE46550
					D3E037E1-3EB8-44C8-A917-57927947596D
					D4F940AB-401B-4EFC-AADC-AD5F3C50688A
		*/

	} else {
		// Windows ASR can not be hardened, since Windows it too old (need at least Windows 10 - 1709)
		return false
	}

	return hardened
}

func (asr WindowsASRStruct) Name() string {
	return asr.shortName
}

func (asr WindowsASRStruct) LongName() string {
	return asr.longName
}

func (asr WindowsASRStruct) Description() string {
	return asr.description
}

// checks if hardentools is running on Windows 10 with Patch Level >= 1709
func checkWindowsVersion() bool {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, "SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	maj, _, err := k.GetIntegerValue("CurrentMajorVersionNumber")
	if err != nil {
		return false
	}
	if maj < 10 {
		return false
	}

	min, _, err := k.GetIntegerValue("CurrentMinorVersionNumber")
	if err != nil {
		return false
	}
	if min < 0 {
		return false
	}

	cb, _, err := k.GetStringValue("CurrentBuild")
	if err != nil {
		return false
	}
	if strings.Compare(cb, "15254") < 0 {
		return false
	}

	return true
}

// helper method for executing powershell commands
func executeCommand(cmd string, args ...string) (string, error) {
	var out []byte
	command := exec.Command(cmd, args...)
	out, err := command.CombinedOutput()

	return string(out), err
}