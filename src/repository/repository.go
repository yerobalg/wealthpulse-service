package repository

import (
	"github.com/yerobalg/wealthpulse-service/helper/db"
	"github.com/yerobalg/wealthpulse-service/helper/httpclient"
)

type Repository struct {
	User       UserInterface
	Role       RoleInterface
	Permission PermissionInterface
	AssetPrice AssetPriceInterface
}

type InitParam struct {
	DB           db.DB
	HTTPClient   httpclient.Interface
	CoinGecko    CoinGeckoConfig
	YahooFinance YahooFinanceConfig
	ExchangeRate ExchangeRateConfig
}

func Init(param InitParam) *Repository {
	return &Repository{
		User:       InitUser(param.DB),
		Role:       InitRole(param.DB),
		Permission: InitPermission(param.DB),
		AssetPrice: InitAssetPrice(
			param.HTTPClient,
			param.CoinGecko,
			param.YahooFinance,
			param.ExchangeRate,
		),
	}
}
