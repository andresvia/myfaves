package main

import (
	"crypto/x509"
	"encoding/csv"
	"fmt"
	"golang.org/x/net/html"
	"gopkg.in/urfave/cli.v1"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"syscall"
	"time"
)

type bookmark struct {
	href        string
	description string
	dir         string
	root        string
	hrefcode    string
	rootcode    string
}

func getcode(geturl string) string {
	client := &http.Client{}
	ch := make(chan string)
	go func() {
		for {
			req, _ := http.NewRequest("HEAD", geturl, nil)
			req.Header.Add("Connection", "close")
			resp, httperr := client.Do(req)
			if neterr, ok := httperr.(net.Error); ok {
				if neterr.Temporary() {
					time.Sleep(time.Nanosecond)
					continue
				} else {
					httperr = neterr
				}
			}
			if urlerr, ok := httperr.(*url.Error); ok {
				if urlerr.Temporary() {
					time.Sleep(time.Nanosecond)
					continue
				} else {
					httperr = urlerr.Err
				}
			}
			if operr, ok := httperr.(*net.OpError); ok {
				httperr = operr.Err
			}
			if x509err, ok := httperr.(x509.SystemRootsError); ok {
				httperr = x509err.Err
			}
			if patherr, ok := httperr.(*os.PathError); ok {
				httperr = patherr.Err
			}
			if dnsErr, ok := httperr.(*net.DNSError); ok {
				if dnsErr.Temporary() || strings.Contains(dnsErr.Err, "oo many open files") || strings.Contains(dnsErr.Err, "esource temporarily unavailable") || strings.Contains(dnsErr.Err, "emporary failure in name resolution") || strings.Contains(dnsErr.Err, "connection refused") {
					time.Sleep(time.Nanosecond)
					continue
				}
			}
			if errno, ok := httperr.(syscall.Errno); ok {
				if errno.Temporary() {
					time.Sleep(time.Nanosecond)
					continue
				}
			}
			if httperr != nil {
				httperrValue := reflect.ValueOf(httperr)
				httperrType := fmt.Sprintf(" - %v", httperrValue.Type())
				ch <- httperr.Error() + httperrType
				break
			} else {
				ch <- resp.Status
				break
			}
		}
	}()
	select {
	case str := <-ch:
		return str
	case <-time.After(timeout):
		return timeout.String() + " timeout"
	}
	return <-ch
}

func (b *bookmark) setRoot() {
	u, err := url.Parse(b.href)
	if err != nil {
		u = &url.URL{Scheme: err.Error()}
	} else {
		u.Path = "/"
		u.RawPath = ""
		u.RawQuery = ""
		u.Fragment = ""
	}
	b.root = u.String()
}

var app *cli.App

type bp struct {
	*bookmark
}

var bookmarks []bp = []bp{}

var timeout = time.Minute

func init() {
	app = cli.NewApp()
	app.Action = action
}

func main() {
	app.Run(os.Args)
}

func action(ctx *cli.Context) error {
	for _, arg := range ctx.Args() {
		file, err := os.Open(arg)
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		root, err := html.Parse(file)
		if err != nil {
			return cli.NewExitError(err, 2)
		}
		bookmarkify(root)
		w := csv.NewWriter(os.Stdout)
		resolveAll()
		w.WriteAll(tablify())
	}
	return nil
}

func bookmarkify(n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "a" {
		href := ""
		text := ""
		if n.FirstChild != nil {
			text = n.FirstChild.Data
		}
		for _, a := range n.Attr {
			if a.Key == "href" {
				href = a.Val
				break
			}
		}
		bookmarks = append(bookmarks, bp{bookmark: &bookmark{href, text, n.Parent.Parent.Parent.FirstChild.FirstChild.Data, "", "", ""}})
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		bookmarkify(c)
	}
}

func resolveAll() {
	ch := make(chan int)
	for _, bm := range bookmarks {
		go func(bm *bookmark) {
			bm.setRoot()
			bm.rootcode = getcode(bm.root)
			ch <- 1
		}(bm.bookmark)
		go func(bm *bookmark) {
			bm.hrefcode = getcode(bm.href)
			ch <- 1
		}(bm.bookmark)
	}
	ticker := time.NewTicker(time.Second)
	go func() {
		for {
			<-ticker.C
			ch <- 0
		}
	}()
	mustget := 2 * len(bookmarks)
	chars := []string{"◜ ", " ◝", " ◞", "◟ "}
	then := time.Now()
	for got := <-ch; got < mustget; got += <-ch {
		print(fmt.Sprintf("\r%s - %d/%d - %d/%d", chars[got%len(chars)], got + 1, mustget, time.Since(then)/time.Second, timeout/time.Second))
	}
	ticker.Stop()
	close(ch)

}

func tablify() [][]string {
	table := [][]string{}
	for _, b := range bookmarks {
		row := []string{b.href, b.description, b.dir, b.root, b.hrefcode, b.rootcode}
		table = append(table, row)
	}
	return table
}
