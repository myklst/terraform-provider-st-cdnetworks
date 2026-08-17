package utils

import "github.com/hashicorp/terraform-plugin-framework/types"

func ToStrSlice(list types.List) *[]string {
	if list.IsNull() || list.IsUnknown() || len(list.Elements()) == 0 {
		return &[]string{}
	}
	var out []string
	list.ElementsAs(nil, &out, false)
	return &out
}
