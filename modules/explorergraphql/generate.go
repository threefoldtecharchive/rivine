package explorergraphql

//go:generate go run github.com/99designs/gqlgen -v
//go:generate rm -rf server

//go:generate sh -c "graphql-markdown schema.graphql > schema.md"
