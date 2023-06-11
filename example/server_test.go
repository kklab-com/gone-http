package example

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"math/rand"
	"net"
	http2 "net/http"
	"testing"
	"time"

	"github.com/kklab-com/gone-core/channel"
	"github.com/kklab-com/gone-http/http"
	httpheadername "github.com/kklab-com/gone-httpheadername"
	buf "github.com/kklab-com/goth-bytebuf"
	concurrent "github.com/kklab-com/goth-concurrent"
	"github.com/kklab-com/goth-kklogger"
	"github.com/stretchr/testify/assert"
)

func gUnzipData(data []byte) (resData []byte, err error) {
	b := bytes.NewBuffer(data)

	var r io.Reader
	r, err = gzip.NewReader(b)
	if err != nil {
		return
	}

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return
	}

	resData = resB.Bytes()

	return
}

func TestServer_Start(t *testing.T) {
	kklogger.SetLogLevel("TRACE")
	bootstrap := channel.NewServerBootstrap()
	bootstrap.ChannelType(&http.ServerChannel{})
	bootstrap.Handler(channel.NewInitializer(func(ch channel.Channel) {
		ch.Pipeline().AddLast("INDICATE_HANDLER_INBOUND", &channel.IndicateHandlerInbound{})
		ch.Pipeline().AddLast("INDICATE_HANDLER_OUTBOUND", &channel.IndicateHandlerOutbound{})
	}))

	clientCountHandler := &ServerChildCountHandler{}
	bootstrap.ChildHandler(channel.NewInitializer(func(ch channel.Channel) {
		ch.Pipeline().AddLast("CLIENT_COUNT_HANDLER", clientCountHandler)
		ch.Pipeline().AddLast("INDICATE_HANDLER_INBOUND", &channel.IndicateHandlerInbound{})
		ch.Pipeline().AddLast("NET_STATUS_INBOUND", &channel.NetStatusInbound{})
		ch.Pipeline().AddLast("GZIP_HANDLER", new(http.GZipHandler))
		ch.Pipeline().AddLast("LOG_HANDLER", http.NewLogHandler(false))
		ch.Pipeline().AddLast("DISPATCHER", http.NewDispatchHandler(NewRoute()))
		ch.Pipeline().AddLast("NET_STATUS_OUTBOUND", &channel.NetStatusOutbound{})
		ch.Pipeline().AddLast("INDICATE_HANDLER_OUTBOUND", &channel.IndicateHandlerOutbound{})
	}))

	ch := bootstrap.Bind(&net.TCPAddr{IP: nil, Port: 18080}).Sync().Channel()
	wg := concurrent.WaitGroup{}
	for i := 0; i < 10; i++ {
		go func(i int) {
			wg.Add(1)
			defer func() {
				wg.Done()
			}()

			v := fmt.Sprintf("%d", rand.Int())
			req, _ := http2.NewRequest("GET", fmt.Sprintf("http://localhost:18080/long?v=%s", v), nil)
			req.Header = http2.Header{}
			if i%2 == 0 {
				req.Header.Set(httpheadername.AcceptEncoding, "gzip")
			}

			if rtn, err := http2.DefaultClient.Do(req); err != nil {
				assert.Fail(t, err.Error())
			} else {
				if rtn.Header.Get(httpheadername.ContentEncoding) == "gzip" {
					buffer := bytes.Buffer{}
					buffer.ReadFrom(rtn.Body)
					bufLen := buffer.Len()
					bs, e := gUnzipData(buffer.Bytes())
					if e != nil {
						println(e.Error())
					}

					if bufLen == 0 {
						assert.Fail(t, "len should not be zero")
					}

					assert.Equal(t, 200, rtn.StatusCode)
					assert.EqualValues(t, longMsg+v, string(bs))
				} else {
					assert.EqualValues(t, longMsg+v, string(buf.EmptyByteBuf().WriteReader(rtn.Body).Bytes()))
				}
			}
		}(i)
	}

	for i := 0; i < 10; i++ {
		go func() {
			wg.Add(1)
			if rtn, err := http2.DefaultClient.Get("http://localhost:18080"); err != nil {
				assert.Fail(t, err.Error())
			} else {
				assert.EqualValues(t, "feeling good", string(buf.EmptyByteBuf().WriteReader(rtn.Body).Bytes()))
			}

			wg.Done()
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			wg.Add(1)
			if rtn, err := http2.DefaultClient.Get("http://localhost:18080/home"); err != nil {
				assert.Fail(t, err.Error())
			} else {
				assert.Equal(t, 200, rtn.StatusCode)
				assert.EqualValues(t, "/home", string(buf.EmptyByteBuf().WriteReader(rtn.Body).Bytes()))
			}

			wg.Done()
		}()
	}

	for i := 0; i < 5; i++ {
		go func() {
			wg.Add(1)
			if rtn, err := http2.DefaultClient.Get("http://localhost:18080/v1/home"); err != nil {
				assert.Fail(t, err.Error())
			} else {
				assert.Equal(t, 200, rtn.StatusCode)
				assert.EqualValues(t, "/v1/home", string(buf.EmptyByteBuf().WriteReader(rtn.Body).Bytes()))
			}

			wg.Done()
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			wg.Add(1)
			if rtn, err := http2.DefaultClient.Get("http://localhost:18080/homes"); err != nil {
				assert.Fail(t, err.Error())
			} else {
				assert.EqualValues(t, 404, rtn.StatusCode)
			}

			wg.Done()
		}()
	}

	go func() {
		wg.Add(1)
		if rtn, err := http2.DefaultClient.Get("http://localhost:18080/400"); err != nil {
			assert.Fail(t, err.Error())
		} else {
			assert.EqualValues(t, 400, rtn.StatusCode)
		}

		wg.Done()
	}()

	wg.Wait()

	for i := 0; i < 50; i++ {
		go func() {
			wg.Add(1)
			request, _ := http2.NewRequest("POST", "http://localhost:18080", nil)
			request.Header = http2.Header{}
			request.Header.Set(httpheadername.Authorization, "!!!!")
			if rtn, err := http2.DefaultClient.Do(request); err != nil {
				assert.Fail(t, err.Error())
			} else {
				assert.Equal(t, 200, rtn.StatusCode)
				assert.EqualValues(t, "feeling good", string(buf.EmptyByteBuf().WriteReader(rtn.Body).Bytes()))
			}

			wg.Done()
		}()
	}

	wg.Wait()
	if rtn, err := http2.DefaultClient.Get("http://localhost:18080/close"); err != nil {
		assert.Fail(t, err.Error())
	} else {
		assert.Equal(t, 200, rtn.StatusCode)
		assert.EqualValues(t, "/close", string(buf.EmptyByteBuf().WriteReader(rtn.Body).Bytes()))
	}

	go func() {
		time.Sleep(time.Minute * 5)
		ch.Close()
	}()

	ch.CloseFuture().Sync()
	assert.Equal(t, int32(0), clientCountHandler.regTrigCount)
	assert.Equal(t, int32(0), clientCountHandler.actTrigCount)
}

func TestServer_BodyLimit(t *testing.T) {
	kklogger.SetLogLevel("TRACE")
	bootstrap := channel.NewServerBootstrap()
	bootstrap.ChannelType(&http.ServerChannel{})
	bootstrap.SetParams(http.ParamMaxBodyBytes, 10)
	bootstrap.Handler(channel.NewInitializer(func(ch channel.Channel) {
		ch.Pipeline().AddLast("INDICATE_HANDLER_INBOUND", &channel.IndicateHandlerInbound{})
		ch.Pipeline().AddLast("INDICATE_HANDLER_OUTBOUND", &channel.IndicateHandlerOutbound{})
	}))

	clientCountHandler := &ServerChildCountHandler{}
	bootstrap.ChildHandler(channel.NewInitializer(func(ch channel.Channel) {
		ch.Pipeline().AddLast("CLIENT_COUNT_HANDLER", clientCountHandler)
		ch.Pipeline().AddLast("INDICATE_HANDLER_INBOUND", &channel.IndicateHandlerInbound{})
		ch.Pipeline().AddLast("NET_STATUS_INBOUND", &channel.NetStatusInbound{})
		ch.Pipeline().AddLast("GZIP_HANDLER", new(http.GZipHandler))
		ch.Pipeline().AddLast("LOG_HANDLER", http.NewLogHandler(false))
		ch.Pipeline().AddLast("DISPATCHER", http.NewDispatchHandler(NewRoute()))
		ch.Pipeline().AddLast("NET_STATUS_OUTBOUND", &channel.NetStatusOutbound{})
		ch.Pipeline().AddLast("INDICATE_HANDLER_OUTBOUND", &channel.IndicateHandlerOutbound{})
	}))

	ch := bootstrap.Bind(&net.TCPAddr{IP: nil, Port: 18081}).Sync().Channel()

	wg := concurrent.WaitGroup{}
	wg.Add(1)
	go func() {
		request, _ := http2.NewRequest("POST", "http://localhost:18081", buf.NewByteBufString("this is more than 10 characters"))
		if rtn, err := http2.DefaultClient.Do(request); err != nil {
			assert.Fail(t, "no")
		} else {
			assert.Equal(t, 413, rtn.StatusCode)
		}

		wg.Done()
	}()

	wg.Wait()
	ch.Close()
	ch.CloseFuture().Sync()
}
