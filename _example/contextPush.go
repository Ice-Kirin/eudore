package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	app.GetFunc("/*", func(ctx eudore.Context) {
		ctx.Push("/css/1.css", nil)
		ctx.Push("/css/2.css", nil)
		ctx.Push("/css/3.css", nil)
		ctx.Push("/favicon.ico", nil)
		ctx.WriteString(`<!DOCTYPE html>
<html>
<head>
	<title>push</title>
	<link href='/css/1.css' rel="stylesheet">
	<link href='/css/2.css' rel="stylesheet">
	<link href='/css/3.css' rel="stylesheet">
</head>
<body>
push test
</body>
</html>`)
	})
	app.GetFunc("/css/*", func(ctx eudore.Context) {
		ctx.WriteString("*{}")
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").Do().CheckStatus(200).Out()
	for client.Next() {
		app.Error(client.Error())
	}

	app.Run()
}
