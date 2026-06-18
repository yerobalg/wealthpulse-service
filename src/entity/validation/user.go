package validation

import (
	"github.com/yerobalg/wealthpulse-service/helper/validator"
)

var userCommon = []validator.Response{
	validator.Required("name", "Nama pengguna"),
	validator.Max("name", "Nama pengguna", 255),
	validator.Required("username", "Username"),
	validator.Max("username", "Username", 255),
	validator.PrintASCII("username", "Username"),
}

var UserLogin = []validator.Response{
	validator.Required("username", "Username"),
	validator.Required("password", "Password"),
}

var UserCreate = validator.Concat(
	userCommon,
	Password("password", "Password"),
)

var UserUpdate = validator.Concat(
	userCommon,
	validator.RequiredID("id", "ID pengguna"),
)
