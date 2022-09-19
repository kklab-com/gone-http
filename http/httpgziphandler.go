package http

import (
	"compress/gzip"
	"strings"
	"time"

	"github.com/kklab-com/gone-core/channel"
	"github.com/kklab-com/gone-httpheadername"
	buf "github.com/kklab-com/goth-bytebuf"
)

type GZipHandler struct {
	channel.DefaultHandler
	CompressThreshold int
}

func (h *GZipHandler) Added(ctx channel.HandlerContext) {
	if h.CompressThreshold == 0 {
		h.CompressThreshold = 128
	}
}

func (h *GZipHandler) Write(ctx channel.HandlerContext, obj any, future channel.Future) {
	pack := _UnPack(obj)
	if pack == nil {
		ctx.Write(obj, future)
		return
	}

	response := pack.Response
	params := pack.Params
	if response == nil {
		ctx.Write(obj, future)
		return
	}

	if response.body.ReadableBytes() < 128 {
		ctx.Write(obj, future)
		return
	}

	if strings.Contains(response.request.Header().Get(httpheadername.AcceptEncoding), "gzip") {
		st := time.Now()
		response.SetHeader(httpheadername.ContentEncoding, "gzip")
		response.SetBody(h.gzipWrite(response.body))
		params["[gone-http]compress_time"] = time.Now().Sub(st).Nanoseconds()
	}

	ctx.Write(obj, future)
}

func (h *GZipHandler) gzipWrite(buffer buf.ByteBuf) buf.ByteBuf {
	gzBuffer := buf.EmptyByteBuf()
	writer, _ := gzip.NewWriterLevel(gzBuffer, gzip.BestSpeed)
	defer writer.Close()
	writer.Write(buffer.Bytes())
	writer.Flush()
	return gzBuffer
}
