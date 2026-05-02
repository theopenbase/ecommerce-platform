package service

import "errors"

var (
	ErrSpuNotFound      = errors.New("SPU not found")
	ErrSkuNotFound      = errors.New("SKU not found")
	ErrSpuCodeExists    = errors.New("SPU code already exists")
	ErrSkuCodeExists    = errors.New("SKU code already exists")
	ErrCategoryNotFound = errors.New("category not found")
	ErrBrandNotFound    = errors.New("brand not found")
	ErrInvalidStatus    = errors.New("invalid status")
)
