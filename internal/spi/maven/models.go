package maven

type VulnerabilityReport struct {
	ProjectInfo struct {
		ReportDate string `json:"reportDate"`
		//Credits    struct {
		//	RETIREJS string `json:"RETIREJS"`
		//	NPM      string `json:"NPM"`
		//	NVD      string `json:"NVD"`
		//	OSSINDEX string `json:"OSSINDEX"`
		//} `json:"credits"`
		GroupID    string `json:"groupID"`
		Name       string `json:"name"`
		ArtifactID string `json:"artifactID"`
		Version    string `json:"version"`
	} `json:"projectInfo"`
	ReportSchema string `json:"reportSchema"`
	ScanInfo     struct {
		EngineVersion string `json:"engineVersion"`
		DataSource    []struct {
			Name      string `json:"name"`
			Timestamp string `json:"timestamp"`
		} `json:"dataSource"`
	} `json:"scanInfo"`
	Dependencies []Dependency `json:"dependencies"`
}

type Dependency struct {
	Sha1              string   `json:"sha1"`
	FileName          string   `json:"fileName"`
	ProjectReferences []string `json:"projectReferences"`
	Sha256            string   `json:"sha256"`
	VulnerabilityIDs  []struct {
		Confidence string `json:"confidence"`
		ID         string `json:"id"`
		Url        string `json:"url"`
	} `json:"vulnerabilityIds"`
	FilePath          string          `json:"filePath"`
	Description       string          `json:"description"`
	Vulnerabilities   []Vulnerability `json:"vulnerabilities"`
	IsVirtual         bool            `json:"isVirtual"`
	EvidenceCollected struct {
		ProductEvidence []struct {
			Confidence string `json:"confidence"`
			Name       string `json:"name"`
			Source     string `json:"source"`
			Type       string `json:"type"`
			Value      string `json:"value"`
		} `json:"productEvidence"`
		VendorEvidence []struct {
			Confidence string `json:"confidence"`
			Name       string `json:"name"`
			Source     string `json:"source"`
			Type       string `json:"type"`
			Value      string `json:"value"`
		} `json:"vendorEvidence"`
		VersionEvidence []struct {
			Confidence string `json:"confidence"`
			Name       string `json:"name"`
			Source     string `json:"source"`
			Type       string `json:"type"`
			Value      string `json:"value"`
		} `json:"versionEvidence"`
	} `json:"evidenceCollected"`
	Packages []struct {
		Confidence string `json:"confidence"`
		ID         string `json:"id"`
		Url        string `json:"url"`
	} `json:"packages"`
	Md5 string `json:"md5"`
}

type Vulnerability struct {
	Severity   string `json:"severity"`
	Notes      string `json:"notes"`
	References []struct {
		Name   string `json:"name"`
		Source string `json:"source"`
		Url    string `json:"url"`
	} `json:"references"`
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	Source             string   `json:"source"`
	Cvssv2             CVSSv2   `json:"cvssv2"`
	Cvssv3             CVSSv3   `json:"cvssv3"`
	Cwes               []string `json:"cwes"`
	VulnerableSoftware []struct {
		Software struct {
			VersionEndIncluding    string `json:"versionEndIncluding"`
			ID                     string `json:"id"`
			VulnerabilityIDMatched string `json:"vulnerabilityIdMatched"`
		} `json:"software"`
	} `json:"vulnerableSoftware"`
}

type CVSSv2 struct {
	ConfidentialImpact  string  `json:"confidentialImpact"`
	Severity            string  `json:"severity"`
	Score               float64 `json:"score"`
	ExploitabilityScore string  `json:"exploitabilityScore"`
	AccessComplexity    string  `json:"accessComplexity"`
	AvailabilityImpact  string  `json:"availabilityImpact"`
	IntegrityImpact     string  `json:"integrityImpact"`
	ImpactScore         string  `json:"impactScore"`
	Version             string  `json:"version"`
	AccessVector        string  `json:"accessVector"`
	Authenticationr     string  `json:"authenticationr"`
}

type CVSSv3 struct {
	ExploitabilityScore   string  `json:"exploitabilityScore"`
	AvailabilityImpact    string  `json:"availabilityImpact"`
	BaseScore             float64 `json:"baseScore"`
	PrivilegesRequired    string  `json:"privilegesRequired"`
	UserInteraction       string  `json:"userInteraction"`
	Version               string  `json:"version"`
	BaseSeverity          string  `json:"baseSeverity"`
	ConfidentialityImpact string  `json:"confidentialityImpact"`
	AttackComplexity      string  `json:"attackComplexity"`
	Scope                 string  `json:"scope"`
	AttackVector          string  `json:"attackVector"`
	IntegrityImpact       string  `json:"integrityImpact"`
	ImpactScore           string  `json:"impactScore"`
}
