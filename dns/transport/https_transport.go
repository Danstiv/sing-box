package transport

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/sagernet/sing-box/common/tls"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/logger"
	M "github.com/sagernet/sing/common/metadata"

	"golang.org/x/net/http2"
)

var errFallback = E.New("fallback to HTTP/1.1")

type HTTPSTransportWrapper struct {
	http2Transport *http2.Transport
	httpTransport  *http.Transport
	fallback       *atomic.Bool
}

func NewHTTPSTransportWrapper(logger logger.ContextLogger, dialer tls.Dialer, serverAddr M.Socksaddr) *HTTPSTransportWrapper {
	var fallback atomic.Bool
	return &HTTPSTransportWrapper{
		http2Transport: &http2.Transport{
			DialTLSContext: func(ctx context.Context, _, _ string, _ *tls.STDConfig) (net.Conn, error) {
				start := time.Now()
				logger.WarnContext(ctx, "DoH dialing new h2 connection to ", serverAddr)
				tlsConn, err := dialer.DialTLSContext(ctx, serverAddr)
				if err != nil {
					logger.WarnContext(ctx, "DoH h2 dial failed after ", time.Since(start), ": ", err)
					return nil, err
				}
				logger.WarnContext(ctx, "DoH h2 dial ok in ", time.Since(start))
				state := tlsConn.ConnectionState()
				if state.NegotiatedProtocol == http2.NextProtoTLS {
					return tlsConn, nil
				}
				tlsConn.Close()
				fallback.Store(true)
				return nil, errFallback
			},
		},
		httpTransport: &http.Transport{
			DialTLSContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				start := time.Now()
				logger.WarnContext(ctx, "DoH dialing new fallback connection to ", serverAddr)
				conn, err := dialer.DialTLSContext(ctx, serverAddr)
				if err != nil {
					logger.WarnContext(ctx, "DoH fallback dial failed after ", time.Since(start), ": ", err)
					return nil, err
				}
				logger.WarnContext(ctx, "DoH fallback dial ok in ", time.Since(start))
				return conn, nil
			},
		},
		fallback: &fallback,
	}
}

func (h *HTTPSTransportWrapper) RoundTrip(request *http.Request) (*http.Response, error) {
	if h.fallback.Load() {
		return h.httpTransport.RoundTrip(request)
	} else {
		response, err := h.http2Transport.RoundTrip(request)
		if err != nil {
			if errors.Is(err, errFallback) {
				return h.httpTransport.RoundTrip(request)
			}
			return nil, err
		}
		return response, nil
	}
}

func (h *HTTPSTransportWrapper) CloseIdleConnections() {
	h.http2Transport.CloseIdleConnections()
	h.httpTransport.CloseIdleConnections()
}

func (h *HTTPSTransportWrapper) Clone() *HTTPSTransportWrapper {
	return &HTTPSTransportWrapper{
		httpTransport: h.httpTransport,
		http2Transport: &http2.Transport{
			DialTLSContext: h.http2Transport.DialTLSContext,
		},
		fallback: h.fallback,
	}
}
