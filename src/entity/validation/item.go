package validation

import (
	"github.com/yerobalg/wealthpulse-service/helper/validator"
)

var itemCommon = []validator.Response{
	validator.Required("name", "Nama item"),
	validator.Max("name", "Nama item", 255),
	validator.Max("description", "Deskripsi", 1000),
	validator.GTE("price", "Harga"),
}

var ItemCreate = itemCommon

var ItemUpdate = validator.Concat(
	itemCommon,
	validator.RequiredID("id", "ID item"),
)
