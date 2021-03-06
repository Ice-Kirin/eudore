package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	// 参数是压缩等级
	app.AddMiddleware(middleware.NewGzipFunc(5))
	app.Listen(":8088")
	app.Run()
}
