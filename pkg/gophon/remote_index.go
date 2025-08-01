package gophon

type RemoteIndex struct {
	GitHubOwner string
	GitHubRepo  string
	BaseUrl     string
}

var RemoteIndexMap = map[string]RemoteIndex{
	"github.com/hashicorp/terraform-provider-azurerm/internal": {
		GitHubOwner: "lonegunmanb",
		GitHubRepo:  "terraform-provider-azurerm-index",
		BaseUrl:     "https://raw.githubusercontent.com/lonegunmanb/terraform-provider-azurerm-index/refs/heads/main/index/internal",
	},
}
