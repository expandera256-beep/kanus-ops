package main

import (
    "fmt"
    "testing"
)
func TestChooseInstallLocationUsesAvailableAlternative(t *testing.T) {
    choices:=[]cloudInstallLocation{{parent:"C/US",probe:"C"},{parent:"D/US",probe:"D"}}
    got,err:=chooseCloudInstallLocation(choices,func(path string)(uint64,error){
        if path=="C" { return 123,nil }; return minCloudInstallFreeBytes+1,nil
    })
    if err!=nil || got.parent!="D/US" { t.Fatal(got,err) }
}
func TestChooseInstallLocationRejectsFullOrUnreadableVolumes(t *testing.T) {
    choices:=[]cloudInstallLocation{{parent:"C/US",probe:"C"},{parent:"D/US",probe:"D"}}
    _,err:=chooseCloudInstallLocation(choices,func(path string)(uint64,error){
        if path=="C" { return 0,fmt.Errorf("unreadable") }; return minCloudInstallFreeBytes-1,nil
    })
    if err==nil { t.Fatal("expected insufficient-space error") }
}
