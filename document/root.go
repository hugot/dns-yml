package document

type Root struct {
	Domains     map[string]Domain     `yaml:"domains"`
	RoundRobins map[string]RoundRobin `yaml:"round_robins"`
}

type Domain struct {
	Records []Record `yaml:"records"`
}

type Record struct {
	Type     string        `yaml:"type"`
	Name     string        `yaml:"name"`
	Content  RecordContent `yaml:"content"`
	Priority int           `yaml:"priority"`
	TTL      int           `yaml:"ttl"`
}

type RecordContent struct {
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
}

type RoundRobin []string
