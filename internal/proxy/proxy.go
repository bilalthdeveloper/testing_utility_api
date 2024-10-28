package proxy

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/bilalthdeveloper/kadrion/utils"
	"github.com/gobwas/ws"
	"golang.org/x/exp/rand"
)

type proxyService interface {
	HasProxies() bool
	GetRandomProxy() string
}

type ProxyService struct {
	ProxyList Proxies
}

type Proxies struct {
	List []string
}

func Initialize() *ProxyService {
	proxyService := &ProxyService{
		ProxyList: Proxies{},
	}

	if len(os.Args) > 6 && os.Args[6] != "" {
		proxies, err := Setup(os.Args[6])
		if err != nil {
			utils.LogMessage(err.Error(), 1)
			return proxyService
		}
		proxyService.ProxyList.List = proxies
	}

	return proxyService
}

func Setup(filename string) ([]string, error) {
	var proxies []string
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		proxy := scanner.Text()
		if proxy != "" {
			proxies = append(proxies, proxy)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return proxies, nil
}

func (ps *ProxyService) HasProxies() bool {
	return len(ps.ProxyList.List) > 0
}
func (ps *ProxyService) GetRandomProxy() string {
	rand.Seed(uint64(time.Now().UnixNano()))

	randomIndex := rand.Intn(len(ps.ProxyList.List))

	randomValue := ps.ProxyList.List[randomIndex]
	return randomValue
}

func (ps *ProxyService) GetWsConn(ctx context.Context, proxyStr string, wsURL string) (net.Conn, error) {

	proxyURL, err := url.Parse(proxyStr)
	if err != nil {
		return nil, err
	}

	dialer := ws.Dialer{
		NetDial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", proxyURL.Host)
			if err != nil {
				return nil, err
			}

			req := &http.Request{
				Method: "GET",
				URL:    &url.URL{Host: addr},
				Host:   addr,
			}

			err = req.Write(conn)
			if err != nil {
				conn.Close()
				return nil, err
			}

			resp, err := http.ReadResponse(bufio.NewReader(conn), req)
			if err != nil {
				conn.Close()
				return nil, err
			}
			if resp.StatusCode != http.StatusOK {
				conn.Close()
				return nil, fmt.Errorf("proxy connection failed with status: %s", resp.Status)
			}

			return conn, nil
		},
	}

	conn, _, _, err := dialer.Dial(ctx, wsURL)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
