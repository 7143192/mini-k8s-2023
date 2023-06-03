package defines

// the start character should be upper-case when declaring element in one structure!!!
// const types.

const POD = 1
const NODE = 2
const SERVICE = 3
const REPLICASET = 4
const AUTO = 5
const DEPLOYMENT = 6
const DNS = 7
const GPU = 8

type YamlMetadata struct {
	metadata string
}

type YamlSpec struct {
	spec string
}

type YamlNode struct {
	ApiVersion string
	Kind       string
	Metadata   YamlMetadata
	Spec       YamlSpec
}

// used to read the start part of one config yaml file.

type YamlStart struct {
	ApiVersion string
	Kind       string
}
