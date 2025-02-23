package http

import (
	"github.com/kklab-com/gone-core/channel"
	"net/http"
)

type Channel struct {
	channel.DefaultNetChannel
}

func (c *Channel) UnsafeIsAutoRead() bool {
	return false
}

func (c *Channel) UnsafeRead() (any, error) {
	return nil, nil
}

func (c *Channel) UnsafeWrite(obj any) error {
	pack := _UnPack(obj)
	if pack == nil {
		return channel.ErrUnknownObjectType
	}

	response := pack.Response
	if !response.headerWritten {
		for key, values := range response.header {
			pack.Writer.Header().Del(key)
			for _, value := range values {
				pack.Writer.Header().Add(key, value)
			}
		}

		for _, value := range response.cookies {
			for _, cookie := range value {
				http.SetCookie(pack.Writer, &cookie)
			}
		}

		pack.Writer.WriteHeader(response.statusCode)
	}

	if pack.writeSeparateMode && !response.headerWritten {
		response.headerWritten = true
	} else {
		_, err := pack.Writer.Write(response.Body().Bytes())
		if flusher, ok := pack.Writer.(http.Flusher); ok {
			flusher.Flush()
		}

		return err
	}

	return nil
}
