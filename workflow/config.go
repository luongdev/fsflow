package workflow

type Config struct {
	CadenceHost       string `yaml:"cadenceHost"`
	CadencePort       uint16 `yaml:"cadencePort"`
	CadenceTaskList   string `yaml:"cadenceTaskList"`
	CadenceClientName string `yaml:"cadenceClientName"`
	CadenceService    string `yaml:"cadenceService"`
}
