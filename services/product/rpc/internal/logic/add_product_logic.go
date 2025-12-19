package logic

import (
	"context"
	"strings"
	"time"

	"letsgo/common/errorx"
	"letsgo/services/product/model"
	"letsgo/services/product/rpc/internal/svc"
	"letsgo/services/product/rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddProductLogic {
	return &AddProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Add new product (admin)
func (l *AddProductLogic) AddProduct(in *product.AddProductRequest) (*product.AddProductResponse, error) {
	// 1. Validate input parameters
	if err := l.validateAddProductParams(in); err != nil {
		return nil, err
	}

	// 2. Prepare product data
	now := time.Now().Unix()
	newProduct := &model.Product{
		Name:        in.Name,
		Description: in.Description,
		Price:       in.Price,
		Stock:       in.Stock,
		Category:    in.Category,
		Images:      in.Images,
		Attributes:  in.Attributes,
		Sales:       0, // Initial sales count is 0
		Status:      1, // 1 = active
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 3. Insert product into database
	result, err := l.svcCtx.ProductModel.Insert(l.ctx, newProduct)
	if err != nil {
		l.Logger.Errorf("Failed to insert product: %v", err)
		return nil, errorx.ErrDatabase
	}

	productId, err := result.LastInsertId()
	if err != nil {
		l.Logger.Errorf("Failed to get last insert id: %v", err)
		return nil, errorx.ErrDatabase
	}

	l.Logger.Infof("Product added successfully: product_id=%d, name=%s", productId, in.Name)

	err = IncCategoryVersion(l.ctx, in.Category, &l.svcCtx.Redis)
	if err != nil {
		l.Logger.Errorf("Increase category version failed! err:%s", err)
	}

	err = IncGlobalVersion(l.ctx, &l.svcCtx.Redis)
	if err != nil {
		l.Logger.Errorf("Increase global version failed! err:%s", err)
	}

	return &product.AddProductResponse{
		ProductId: productId,
	}, nil
}

// validateAddProductParams validates add product input parameters
func (l *AddProductLogic) validateAddProductParams(in *product.AddProductRequest) error {
	// Validate name
	if len(strings.TrimSpace(in.Name)) == 0 {
		return errorx.NewCodeError(1001, "Product name cannot be empty")
	}
	if len(in.Name) > 200 {
		return errorx.NewCodeError(1001, "Product name must be less than 200 characters")
	}

	// Validate price
	if in.Price <= 0 {
		return errorx.NewCodeError(1001, "Product price must be greater than 0")
	}

	// Validate stock
	if in.Stock < 0 {
		return errorx.NewCodeError(1001, "Product stock cannot be negative")
	}

	// Validate category
	if len(strings.TrimSpace(in.Category)) == 0 {
		return errorx.NewCodeError(1001, "Product category cannot be empty")
	}

	return nil
}
