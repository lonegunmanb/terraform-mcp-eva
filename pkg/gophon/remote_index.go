package gophon

type RemoteIndex struct {
	GitHubOwner string
	GitHubRepo  string
	PackagePath string
}

var ProviderIndexMap = map[string]string{
	"azurerm": AzureRMInternal,
	"azuread": AzureADInternal,
	"aws":     AWSInternal,
}

const (
	AzureRMInternal = "github.com/hashicorp/terraform-provider-azurerm/internal"
	AzureADInternal = "github.com/hashicorp/terraform-provider-azuread/internal"
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
		PackagePath: "github.com/hashicorp/terraform-provider-azurerm",
	},
	AzureADInternal: {
		GitHubOwner: "lonegunmanb",
		GitHubRepo:  "terraform-provider-azuread-index",
		PackagePath: "github.com/hashicorp/terraform-provider-azuread",
	},
	AWSInternal: {
		GitHubOwner: "lonegunmanb",
		GitHubRepo:  "terraform-provider-aws-index",
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
