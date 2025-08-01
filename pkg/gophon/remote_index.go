package gophon

type RemoteIndex struct {
	GitHubOwner string
	GitHubRepo  string
	BaseUrl     string
}

var ProviderIndexMap = map[string]string{
	"azurerm": AzureRMInternal,
}

const (
	AzureRMInternal = "github.com/hashicorp/terraform-provider-azurerm/internal"
)

var RemoteIndexMap = map[string]RemoteIndex{
	AzureRMInternal: {
		GitHubOwner: "lonegunmanb",
		GitHubRepo:  "terraform-provider-azurerm-index",
		BaseUrl:     "https://raw.githubusercontent.com/lonegunmanb/terraform-provider-azurerm-index/refs/heads/main/index/internal",
	},
}

// GetSupportedProviders returns a slice of all supported provider names
func GetSupportedProviders() []string {
	providers := make([]string, 0, len(ProviderIndexMap))
	for providerName := range ProviderIndexMap {
		providers = append(providers, providerName)
	}
	return providers
}
