package eudore

import (
	"fmt"
	"encoding/json"
	"encoding/xml"
)

type (
	Renderer interface {
		Render(ResponseWriter, interface{}) error
		ContentType() string
	}
	rendererText struct {}
	rendererJson struct {}
	rendererIndentJson struct {}
	rendererXml struct {}
)

var (
	RendererText		=	rendererText{}
	RendererJson		=	rendererJson{}
	RendererIndentJson		=	rendererIndentJson{}
	RendererXml			=	rendererXml{}
)



func (rendererText) Render(w ResponseWriter, i interface{}) error {
	_, err := fmt.Fprint(w, i)
	return err
}

func (rendererText) ContentType() string {
	const textContentType = MimeTextPlainCharsetUtf8
	return textContentType
}


func (rendererJson) Render(w ResponseWriter, i interface{}) error {
	return json.NewEncoder(w).Encode(i)
}

func (rendererJson) ContentType() string {
	const jsonContentType = MimeApplicationJsonUtf8
	return jsonContentType
}

func (rendererIndentJson) Render(w ResponseWriter, i interface{}) error {
	en := json.NewEncoder(w)
	en.SetIndent("", "\t")
	return en.Encode(i)
}

func (rendererIndentJson) ContentType() string {
	const jsonContentType = MimeApplicationJsonUtf8
	return jsonContentType
}

func (rendererXml) Render(w ResponseWriter, i interface{}) error {
	return xml.NewEncoder(w).Encode(i)
}

func (rendererXml) ContentType() string {
	const xmlContentType = MimeApplicationxmlCharsetUtf8
	return xmlContentType
}
