package main

import (
	"encoding/xml"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func errResponse(ctx *gin.Context, err error) {
	setContentType(ctx.Writer, plainContentType)
	ctx.Writer.WriteHeader(http.StatusUnauthorized)
	ctx.Writer.WriteString(err.Error())
}

func stringResponse(ctx *gin.Context, str string) {
	ctx.Writer.WriteString(str)
}

func xmlResponse(ctx *gin.Context, resp any) {
	bytes, err := xml.Marshal(resp)
	if err != nil {
		log.Errorf("marshal response error: %s", err.Error())
		errResponse(ctx, errors.Wrap(err, "marshal response error"))
		return
	}

	setContentType(ctx.Writer, xmlContentType)
	ctx.Writer.Write(bytes)
}
