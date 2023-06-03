package defines

import "sync"

const WorkFlowPrefix = "WorkFlow"

type Para struct {
	Name string `yaml:"name" json:"name"`
	Kind string `yaml:"kind" json:"kind"`
}

type Task struct {
	Name   string `yaml:"name" json:"name"`
	Params []Para `yaml:"params" json:"params"`
}

type Choice struct {
	Name      string    `yaml:"name" json:"name"`
	Condition Condition `yaml:"condition" json:"condition"`
	True      string    `yaml:"true" json:"true"`
	False     string    `yaml:"false" json:"false"`
}

type Condition struct {
	ParamL Para   `yaml:"L" json:"L"`
	ParamR Para   `yaml:"R" json:"R"`
	Symbol string `yaml:"symbol" json:"symbol"`
}

type Relationship struct {
	Left  string `yaml:"L" json:"L"`
	Right string `yaml:"R" json:"R"`
}

type WorkFlow struct {
	Kind          string         `yaml:"kind" json:"kind"`
	Name          string         `yaml:"name" json:"name"`
	Start         string         `yaml:"start" json:"start"`
	Tasks         []Task         `yaml:"tasks" json:"tasks"`
	Choices       []Choice       `yaml:"choices" json:"choices"`
	Relationships []Relationship `yaml:"relationships" json:"relationships"`
	Lock          sync.RWMutex
}
