// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package product

import (
	"context"

	"letsgo/gateway/internal/svc"
	"letsgo/gateway/internal/types"
	"letsgo/services/product/rpc/product_client"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Add product - Admin creates new product (admin only)
func NewAddProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddProductLogic {
	return &AddProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddProductLogic) AddProduct(req *types.AddProductReq) (resp *types.AddProductResp, err error) {
	ProductResp, err := l.svcCtx.ProductRpc.AddProduct(l.ctx, &product_client.AddProductRequest{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
		Images:      req.Images,
		Attributes:  req.Attributes,
	})
	if err != nil {
		return nil, err
	}

	return &types.AddProductResp{
		ProductId: ProductResp.ProductId,
	}, nil
}
