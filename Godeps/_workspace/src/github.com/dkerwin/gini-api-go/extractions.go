package giniapi

// Box struct
type Box struct {
	Height float64 `json:"height"`
	Left   float64 `json:"left"`
	Page   int     `json:"page"`
	Top    float64 `json:"top"`
	Width  float64 `json:"width"`
}

// Extraction struct
type Extraction struct {
	Box        `json:"box"`
	Candidates string `json:"candidates"`
	Entity     string `json:"entity"`
	Value      string `json:"value"`
}

// Document extractions struct
type Extractions struct {
	Candidates  map[string][]Extraction `json:"candidates"`
	Extractions map[string]Extraction   `json:"extractions"`
}

// GetValue is a helper function to get the extraction value or a empty string
func (e *Extractions) GetValue(key string) string {
	if val, ok := e.Extractions[key]; ok {
		return val.Value
	}
	return ""
}
