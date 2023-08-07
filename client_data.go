package polygon

type ListPackagesInput struct {
	ProblemID int
}

type DownloadPackageInput struct {
	ProblemID int
	PackageID int
	Type      string
}

type Package struct {
	ID                  int    `json:"id"`                  // package's id
	Revision            int    `json:"revision"`            // revision of the problem for which the package was created
	CreationTimeSeconds int    `json:"creationTimeSeconds"` // creation time in unix format
	State               string `json:"state"`               // PENDING/RUNNING/READY/FAILED
	Comment             string `json:"comment"`             // comment for the package
	Type                string `json:"type"`                // type of the package: standard/linux/windows
}
