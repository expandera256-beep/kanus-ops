package main

import "fmt"

// A download and extraction coexist temporarily; keep roughly 1 GiB free.
const minCloudInstallFreeBytes uint64 = 1 << 30

type cloudInstallLocation struct {
	parent string // US-owned folder where the SDK will live
	probe  string // an existing directory on the same volume
}

func chooseCloudInstallLocation(locations []cloudInstallLocation, free func(string) (uint64, error)) (cloudInstallLocation, error) {
	for _, candidate := range locations {
		available, err := free(candidate.probe)
		if err == nil && available >= minCloudInstallFreeBytes {
			return candidate, nil
		}
	}
	return cloudInstallLocation{}, fmt.Errorf("hiçbir yerel sürücüde en az 1 GiB boş alan bulunamadı; disk alanı açın ve Yeniden bağlan'a basın")
}
