package domain

import "errors"

var (
	ErrInvalid      = errors.New("输入不符合桌游规则")
	ErrMissing      = errors.New("记录不存在")
	ErrConflict     = errors.New("对局状态冲突")
	ErrForbidden    = errors.New("没有访问该对局的权限")
	ErrUnauthorized = errors.New("未授权")
	ErrCapacity     = errors.New("席位容量已满")
)
