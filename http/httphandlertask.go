package http

import (
	"fmt"
	"github.com/kklab-com/gone-core/channel"
	httpheadername "github.com/kklab-com/gone-httpheadername"
	"github.com/kklab-com/goth-erresponse"
)

type HttpTask interface {
	Index(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse
	Get(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse
	Create(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse
	Post(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse
	Put(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse
	Delete(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse
	Options(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse
	Patch(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse
	Trace(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse
	Connect(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse
}

type HandlerTask interface {
	GetNodeName(params map[string]any) string
	GetID(name string, params map[string]any) string
}

type HttpHandlerTask interface {
	HttpTask
	CORSHelper(req *Request, resp *Response, params map[string]any)
	PreCheck(req *Request, resp *Response, params map[string]any) ErrorResponse
	Before(req *Request, resp *Response, params map[string]any) ErrorResponse
	After(req *Request, resp *Response, params map[string]any) ErrorResponse
	ErrorCaught(req *Request, resp *Response, params map[string]any, err ErrorResponse) error
}

type SSEOperation interface {
	WriteHeader(ctx channel.HandlerContext, resp *Response, params map[string]any) channel.Future
	WriteBody(ctx channel.HandlerContext, resp *Response, params map[string]any) channel.Future
}

var NotImplemented = erresponse.NotImplemented

type DefaultHTTPHandlerTask struct {
	DefaultHandlerTask
}

func (h *DefaultHTTPHandlerTask) Index(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse {
	return NotImplemented
}

func (h *DefaultHTTPHandlerTask) Get(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse {
	return nil
}

func (h *DefaultHTTPHandlerTask) Create(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse {
	return NotImplemented
}

func (h *DefaultHTTPHandlerTask) Post(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse {
	return nil
}

func (h *DefaultHTTPHandlerTask) Put(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse {
	return nil
}

func (h *DefaultHTTPHandlerTask) Delete(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse {
	return nil
}

func (h *DefaultHTTPHandlerTask) Options(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse {
	return nil
}

func (h *DefaultHTTPHandlerTask) Patch(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse {
	return nil
}

func (h *DefaultHTTPHandlerTask) Trace(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse {
	return nil
}

func (h *DefaultHTTPHandlerTask) Connect(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) ErrorResponse {
	return nil
}

func (h *DefaultHTTPHandlerTask) SSEMode(ctx channel.HandlerContext, req *Request, resp *Response, params map[string]any) SSEOperation {
	if obj, f := params["[gone-http]context_pack"]; f && obj != nil {
		response := obj.(*Pack).Response
		response.SetHeader(httpheadername.ContentType, "text/event-stream")
		response.SetHeader(httpheadername.CacheControl, "no-cache")
		response.SetHeader(httpheadername.Connection, "keep-alive")
		obj.(*Pack).writeSeparateMode = true
		return _DefaultSSEOperation
	}

	return nil
}

var _DefaultSSEOperation = &DefaultSSEOperation{}

type DefaultSSEOperation struct {
}

func (h *DefaultSSEOperation) WriteHeader(ctx channel.HandlerContext, resp *Response, params map[string]any) channel.Future {
	if obj, f := params["[gone-http]context_pack"]; f && obj != nil {
		if dispatcher, f := params["[gone-http]dispatcher"]; f {
			return dispatcher.(*DispatchHandler).callWriteHeader(ctx, obj)
		}
	}

	chCtx := channel.NewFuture(ctx.Channel())
	chCtx.Completable().Fail(fmt.Errorf("not found pack"))
	return chCtx
}

func (h *DefaultSSEOperation) WriteBody(ctx channel.HandlerContext, resp *Response, params map[string]any) channel.Future {
	if obj, f := params["[gone-http]context_pack"]; f && obj != nil {
		pack := obj.(*Pack)
		if !pack.Response.headerWritten {
			if !h.WriteHeader(ctx, resp, params).IsSuccess() {
				chCtx := channel.NewFuture(ctx.Channel())
				chCtx.Completable().Fail(fmt.Errorf("header write error"))
				return chCtx
			}
		}

		return ctx.Write(obj, channel.NewFuture(ctx.Channel()))
	}

	chCtx := channel.NewFuture(ctx.Channel())
	chCtx.Completable().Fail(fmt.Errorf("not found pack"))
	return chCtx
}

func (h *DefaultHTTPHandlerTask) ThrowErrorResponse(err ErrorResponse) {
	panic(err)
}

func (h *DefaultHTTPHandlerTask) CORSHelper(req *Request, resp *Response, params map[string]any) {
	if req.Origin() == "null" {
		resp.Header().Set(httpheadername.AccessControlAllowOrigin, "*")
	} else {
		resp.Header().Set(httpheadername.AccessControlAllowOrigin, req.Origin())
	}

	if str := req.Header().Get(httpheadername.AccessControlRequestHeaders); str != "" {
		resp.Header().Set(httpheadername.AccessControlAllowHeaders, str)
	}

	if str := req.Header().Get(httpheadername.AccessControlRequestMethod); str != "" {
		resp.Header().Set(httpheadername.AccessControlAllowMethods, str)
	}
}

func (h *DefaultHTTPHandlerTask) PreCheck(req *Request, resp *Response, params map[string]any) ErrorResponse {
	return nil
}

func (h *DefaultHTTPHandlerTask) Before(req *Request, resp *Response, params map[string]any) ErrorResponse {
	return nil
}

func (h *DefaultHTTPHandlerTask) After(req *Request, resp *Response, params map[string]any) ErrorResponse {
	return nil
}

func (h *DefaultHTTPHandlerTask) ErrorCaught(req *Request, resp *Response, params map[string]any, err ErrorResponse) error {
	resp.ResponseError(err)
	return nil
}

type DefaultHandlerTask struct {
}

func NewDefaultHandlerTask() *DefaultHandlerTask {
	return new(DefaultHandlerTask)
}

func (h *DefaultHandlerTask) IsIndex(params map[string]any) bool {
	if rtn := params["[gone-http]is_index"]; rtn != nil {
		if is, ok := rtn.(bool); ok && is {
			return true
		}
	}

	return false
}

func (h *DefaultHandlerTask) GetNodeName(params map[string]any) string {
	if rtn := params["[gone-http]node_name"]; rtn != nil {
		return rtn.(string)
	}

	return ""
}

func (h *DefaultHandlerTask) GetID(name string, params map[string]any) string {
	if rtn := params[fmt.Sprintf("[gone-http]%s_id", name)]; rtn != nil {
		return rtn.(string)
	}

	return ""
}

func (h *DefaultHandlerTask) LogExtend(key string, value any, params map[string]any) {
	if rtn := params["[gone-http]extend"]; rtn == nil {
		rtn = map[string]any{key: value}
		params["[gone-http]extend"] = rtn
	} else {
		rtn.(map[string]any)[key] = value
	}
}
