package pricebookentry

import (
	"PowerX/internal/model/product"
	"PowerX/internal/svc"
	"PowerX/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPriceBookEntryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPriceBookEntryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPriceBookEntryLogic {
	return &GetPriceBookEntryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPriceBookEntryLogic) GetPriceBookEntry(req *types.GetPriceBookEntryRequest) (resp *types.GetPriceBookEntryReply, err error) {
	// todo: add your logic here and delete this line

	return
}

func TransformPriceBookEntryToPriceBookEntryReply(entry *product.PriceBookEntry) (entryReply *types.PriceBookEntry) {
	if entry == nil {
		return nil
	}

	discount := CalDiscount(entry.UnitPrice, entry.ListPrice)
	return &types.PriceBookEntry{
		PriceBookId:  entry.PriceBookId,
		ProductId:    entry.ProductId,
		UnitPrice:    entry.UnitPrice,
		ListPrice:    entry.ListPrice,
		Discount:     discount,
		IsActive:     entry.IsActive,
		PriceConfigs: TransformPriceConfigToPriceConfigReply(entry.PriceConfigs),
	}

}

func TransformPriceConfigToPriceConfigReply(config []*product.PriceConfig) (entryReply []*types.PriceConfig) {
	if config == nil {
		return nil
	}
	// to be dev

	return
}
