package main

import (
	"golang.org/x/net/html"
	"gopkg.in/urfave/cli.v1"
	"os"
	"strings"
	"errors"
	"encoding/csv"
)

type bookmark struct {
	href        string
	description string
	tags        []string
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
			return err
		}
		root, err := html.Parse(file)
		if err != nil {
			return err
		}
		err = bookmarkify(root, []string{})
		if err != nil {
			return err
		}
		w := csv.NewWriter(os.Stdout)
		w.WriteAll(tablify())
	}
	return nil
}

func bookmarkify(n *html.Node, tags []string) error {
	if n.Type == html.ElementNode && n.Data == "a" {
		href := ""
		text := ""
		if n.FirstChild != nil && n.FirstChild.Type != html.TextNode {
			return errors.New("a with first child not a document element")
		} else if n.FirstChild != nil {
			text = n.FirstChild.Data
		}
		for _, a := range n.Attr {
			if a.Key == "href" {
				href = a.Val
				break
			}
		}
		bookmarks = append(bookmarks, bookmark{href, text, tags})
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		bookmarkify(c, append(tags, n.Data))
	}
	return nil
}

func tablify() [][]string {
	table := [][]string{}
	for _, b := range bookmarks {
		tags := strings.Join(b.tags, ",")
		row := []string{b.href, b.description, tags}
		table = append(table, row)
	}
	return table
}
