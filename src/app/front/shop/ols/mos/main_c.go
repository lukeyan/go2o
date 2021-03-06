/**
 * Copyright 2014 @ z3q.net.
 * name :
 * author : jarryliu
 * date : 2014-02-05 21:53
 * description :
 * history :
 */
package mos

import (
	"github.com/jsix/gof/web"
)

type mainC struct {
	*baseC
}

func (this *mainC) Login(ctx *web.Context) {
	_, w := ctx.Request, ctx.Response
	ctx.App.Template().Execute(w, nil, "views/ucenter/{device}/login.html")
}

func (this *mainC) Index(ctx *web.Context) {
	if this.Requesting(ctx) {
		_, w := ctx.Request, ctx.Response
		p := this.GetPartner(ctx)
		w.Write([]byte(p.Name))
	}
}
