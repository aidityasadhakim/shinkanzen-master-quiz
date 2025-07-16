package internal

import (
	"embed"
	"encoding/json"
)

type GrammarPoint struct {
	Explanation    string `json:"explanation"`
	FormationRule  string `json:"formation_rule"`
	GrammarPattern string `json:"grammar_pattern"`
	PointID        string `json:"point_id"`
}

type GrammarData struct {
	GrammarPoints []GrammarPoint `json:"grammar_points"`
}

func ReadGrammarData(filename string, templateFS embed.FS) (*GrammarData, error) {
	file, err := templateFS.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var grammarData GrammarData
	if err := json.NewDecoder(file).Decode(&grammarData); err != nil {
		return nil, err
	}
	return &grammarData, nil
}
