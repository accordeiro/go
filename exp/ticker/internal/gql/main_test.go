package gql

import (
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/stellar/go/exp/ticker/internal/gql/schema"
)

func TestValidateSchema(t *testing.T) {
	r := resolver{}
	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	graphql.MustParseSchema(schema.String(), &r, opts...)
}
