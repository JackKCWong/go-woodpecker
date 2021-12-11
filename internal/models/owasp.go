package models

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
	Dependencies []struct {
		Sha1              string   `json:"sha1"`
		License           string   `json:"license"`
		FileName          string   `json:"fileName"`
		ProjectReferences []string `json:"projectReferences"`
		Sha256            string   `json:"sha256"`
		VulnerabilityIDs  []struct {
			Confidence string `json:"confidence"`
			ID         string `json:"id"`
		} `json:"vulnerabilityIds"`
		FilePath          string `json:"filePath"`
		Description       string `json:"description"`
		IsVirtual         bool   `json:"isVirtual"`
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
	} `json:"dependencies"`
}
