package main

import (
	"context"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

type Article struct {
	db *weaviate.Client
}

func (a *Article) Create(ctx context.Context, title, summary string) error {
	_, err := a.db.Data().Creator().
		WithClassName("Article").
		WithProperties(map[string]interface{}{
			"title":   title,
			"summary": summary,
		}).
		Do(ctx)

	return err
}

func (a *Article) CreateSchema(ctx context.Context) error {

	if err := a.db.Schema().AllDeleter().Do(ctx); err != nil {
		return err
	}

	return a.db.Schema().ClassCreator().WithClass(&models.Class{
		Class: "Article",
		ModuleConfig: map[string]map[string]interface{}{
			"qna-openai": {
				"model":            "text-davinci-003",
				"maxTokens":        128,
				"temperature":      0.0,
				"topP":             1,
				"frequencyPenalty": 0.0,
				"presencePenalty":  0.0,
			},
		},
	}).Do(ctx)
}

func (a *Article) Ask(ctx context.Context, question string) (string, error) {
	result, err := a.db.GraphQL().Get().
		WithClassName("Article").
		WithFields(
			graphql.Field{Name: "title"},
			graphql.Field{Name: "_additional", Fields: []graphql.Field{
				{Name: "answer", Fields: []graphql.Field{
					{Name: "hasAnswer"},
					{Name: "property"},
					{Name: "result"},
					{Name: "startPosition"},
					{Name: "endPosition"},
				}},
			}},
		).
		WithAsk(a.db.GraphQL().AskArgBuilder().
			WithQuestion(question).
			WithProperties([]string{"summary"})).
		WithLimit(1).
		Do(ctx)

	if err != nil {
		return "", err
	}

	raw, err := result.MarshalBinary()
	if err != nil {
		return "", err
	}

	fmt.Println(string(raw))

	return gjson.Get(string(raw), "data.Get.Article.0._additional.answer.result").String(), err
}
