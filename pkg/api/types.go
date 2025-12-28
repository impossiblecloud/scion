package api

import (
	"context"
)

type AgentConfig struct {
	Grove      string            `json:"grove"`
	Name       string            `json:"name"`
	Status     string            `json:"status,omitempty"`
	Kubernetes *AgentK8sMetadata `json:"kubernetes,omitempty"`
}

type AgentK8sMetadata struct {
	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	PodName   string `json:"podName"`
	SyncedAt  string `json:"syncedAt,omitempty"`
}

type VolumeMount struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	ReadOnly bool   `json:"read_only,omitempty"`
}

type KubernetesConfig struct {
	Context          string        `json:"context,omitempty"`
	Namespace        string        `json:"namespace,omitempty"`
	RuntimeClassName string        `json:"runtimeClassName,omitempty"`
	Resources        *K8sResources `json:"resources,omitempty"`
}

type K8sResources struct {
	Requests map[string]string `json:"requests,omitempty"`
	Limits   map[string]string `json:"limits,omitempty"`
}

type ScionConfig struct {
	Template        string            `json:"template"`
	HarnessProvider string            `json:"harness_provider,omitempty"`
	ConfigDir       string            `json:"config_dir,omitempty"`
	Env             map[string]string `json:"env,omitempty"`
	Volumes         []VolumeMount     `json:"volumes,omitempty"`
	UnixUsername    string            `json:"unix_username"`
	Image           string            `json:"image"`
	Detached        *bool             `json:"detached"`
	UseTmux         *bool             `json:"use_tmux"`
	Model           string            `json:"model"`
	Runtime         string            `json:"runtime,omitempty"`
	Kubernetes      *KubernetesConfig `json:"kubernetes,omitempty"`
	Agent           *AgentConfig      `json:"agent,omitempty"`
}

func (c *ScionConfig) IsDetached() bool {
	if c.Detached == nil {
		return true
	}
	return *c.Detached
}

func (c *ScionConfig) IsUseTmux() bool {
	if c.UseTmux == nil {
		return false
	}
	return *c.UseTmux
}

type AuthConfig struct {
	GeminiAPIKey         string
	GoogleAPIKey         string
	VertexAPIKey         string
	GoogleAppCredentials string
	GoogleCloudProject   string
	OAuthCreds           string
	AnthropicAPIKey      string
}

// Harness interface moved from pkg/harness to avoid cycles
type Harness interface {
	Name() string
	DiscoverAuth(agentHome string) AuthConfig
	GetEnv(agentName string, unixUsername string, model string, auth AuthConfig) map[string]string
	GetCommand(task string, resume bool) []string
	PropagateFiles(homeDir, unixUsername string, auth AuthConfig) error
	GetVolumes(unixUsername string, auth AuthConfig) []VolumeMount
	DefaultConfigDir() string
	HasSystemPrompt() bool
}

type AgentInfo struct {
	ID          string
	Name        string
	Template    string
	Grove       string
	GrovePath   string
	Labels      map[string]string
	Annotations map[string]string
	Status      string // Container status
	AgentStatus string // Scion agent high-level status
	Image       string
}

type RunConfig struct {
	Name         string
	Template     string
	UnixUsername string
	Image        string
	HomeDir      string
	Workspace    string
	RepoRoot     string
	Env          []string
	Volumes      []VolumeMount
	Labels       map[string]string
	Annotations  map[string]string
	Auth         AuthConfig
	Harness      Harness
	UseTmux      bool
	Model        string
	Task         string
	Resume       bool
}

type Runtime interface {
	Run(ctx context.Context, config RunConfig) (string, error)
	Stop(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, labelFilter map[string]string) ([]AgentInfo, error)
	GetLogs(ctx context.Context, id string) (string, error)
	Attach(ctx context.Context, id string) error
	ImageExists(ctx context.Context, image string) (bool, error)
	PullImage(ctx context.Context, image string) error
}
