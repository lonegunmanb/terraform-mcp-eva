package gophon

type RemoteIndex struct {
	GitHubOwner string
	GitHubRepo  string
	BaseUrl     string
	PackagePath string
}

var ProviderIndexMap = map[string]string{
	"azurerm": AzureRMInternal,
	"aws":     AWSInternal,
}

const (
	AzureRMInternal = "github.com/hashicorp/terraform-provider-azurerm/internal"
	AWSInternal     = "github.com/hashicorp/terraform-provider-aws/internal"
)

var Namespaces = func() []string {
	var s []string
	for k, _ := range RemoteIndexMap {
		s = append(s, k)
	}
	return s
}()

var RemoteIndexMap = map[string]RemoteIndex{
	AzureRMInternal: {
		GitHubOwner: "lonegunmanb",
		GitHubRepo:  "terraform-provider-azurerm-index",
		BaseUrl:     "https://raw.githubusercontent.com/lonegunmanb/terraform-provider-azurerm-index/refs/{version}/index",
		PackagePath: "github.com/hashicorp/terraform-provider-azurerm",
	},
	AWSInternal: {
		GitHubOwner: "lonegunmanb",
		GitHubRepo:  "terraform-provider-aws-index",
		BaseUrl:     "https://raw.githubusercontent.com/lonegunmanb/terraform-provider-aws-index/refs/{version}/index",
		PackagePath: "github.com/hashicorp/terraform-provider-aws",
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
