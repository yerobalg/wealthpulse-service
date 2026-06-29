package repository

import (
	"github.com/yerobalg/wealthpulse-service/helper/db"
	"github.com/yerobalg/wealthpulse-service/helper/httpclient"
)

type Repository struct {
	User         UserInterface
	Role         RoleInterface
	Permission   PermissionInterface
	RevokedToken RevokedTokenInterface
	ActivityLog  ActivityLogInterface
	AssetPrice   AssetPriceInterface
	ExchangeRate ExchangeRateInterface
}

type InitParam struct {
	DB           db.DB
	HTTPClient   httpclient.Interface
	CoinGecko    CoinGeckoConfig
	ExchangeRate ExchangeRateConfig
}

func Init(param InitParam) *Repository {
	return &Repository{
		User:         InitUser(param.DB),
		Role:         InitRole(param.DB),
		Permission:   InitPermission(param.DB),
		RevokedToken: InitRevokedToken(param.DB),
		ActivityLog:  InitActivityLog(param.DB),
		AssetPrice:   InitAssetPrice(param.HTTPClient, param.CoinGecko),
		ExchangeRate: InitExchangeRate(param.HTTPClient, param.ExchangeRate),
	}
}
