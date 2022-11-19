package services

import "github.com/wangxuesong/tcpshadow/model"

type (
	Context struct {
		Data      *model.Data
		SessionId int
	}
)

func NewContext(id int, data *model.Data) *Context {
	return &Context{
		Data:      data,
		SessionId: id,
	}
}
