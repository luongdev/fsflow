package workflow

type Config struct {
	Host        string   `yaml:"host"`
	Port        uint16   `yaml:"port"`
	TaskList    string   `yaml:"task_list"`
	ClientName  string   `yaml:"client_name"`
	ServiceName string   `yaml:"service_name"`
	Domains     []string `yaml:"domains"`
}
