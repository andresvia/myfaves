package main

import (
	"encoding/csv"
	"golang.org/x/net/html"
	"gopkg.in/urfave/cli.v1"
	"os"
)

type bookmark struct {
	href        string
	description string
	dir         string
	root string
	hrefcode string
	rootcode string
}

func getcode(url string) string {
	return "getcode not implemented"
}

func getroot(url string) string {
	return "getroot not implemented"
}

var app *cli.App
var bookmarks []bookmark = []bookmark{}

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
		bookmarks = append(bookmarks, bookmark{href, text, n.Parent.Parent.Parent.FirstChild.FirstChild.Data, "", "", ""})
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		bookmarkify(c)
	}
}

func resolveAll() {
	ch := make(chan bool)
	for _, bookmark := range(bookmarks) {
		go func() {
			bookmark.root = getroot(bookmark.href)
			bookmark.rootcode = getcode(bookmark.root)
			ch <- true
		}()
		go func() {
			bookmark.hrefcode = getcode(bookmark.href)
			ch <- true
		}()
	}
	mustget := 2 * len(bookmarks)
	got := 0
	for <- ch {
		got++
		println("got it")
		if(got >= mustget) {
			close(ch)
		}
	}

}

func tablify() [][]string {
	table := [][]string{}
	for _, b := range bookmarks {
		row := []string{b.href, b.description, b.dir}
		table = append(table, row)
	}
	return table
}
