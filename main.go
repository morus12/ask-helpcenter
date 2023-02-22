package main

import (
	"context"
	"errors"

	"github.com/gocolly/colly"
	"github.com/sirupsen/logrus"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/fault"
)

func main() {
	logrus.SetReportCaller(true)
	ctx := context.Background()

	client := weaviate.New(weaviate.Config{
		Host:   "localhost:8080",
		Scheme: "http",
	})

	article := &Article{db: client}

	// comment index after first run
	if err := index(ctx, article); err != nil {
		logrus.Fatal(err)
	}

	ans, err := article.Ask(ctx, "What is HelpDesk?")
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info(ans)
}

func index(ctx context.Context, article *Article) error {

	if err := article.CreateSchema(ctx); err != nil {
		return err
	}

	c := colly.NewCollector(
		colly.AllowedDomains("www.helpdesk.com"),
		colly.Async(false),
	)

	c.OnHTML("article", func(h *colly.HTMLElement) {
		title := h.DOM.Find("h1").Text()
		summary := h.DOM.Find("div.c-hc-article-content").Text()

		err := article.Create(ctx, title, summary)
		var clientErr *fault.WeaviateClientError
		if errors.As(err, &clientErr) {
			logrus.Warn(clientErr.DerivedFromError)
		} else if err != nil {
			logrus.Warn(err)
		}
	})

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if err := c.Visit(e.Request.AbsoluteURL(link)); err != nil {
			logrus.Warn(err)
		}
	})

	return c.Visit("https://www.helpdesk.com/help/")
}
