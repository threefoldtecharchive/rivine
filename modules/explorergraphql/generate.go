package explorergraphql

//go:generate go run github.com/99designs/gqlgen -v
//go:generate rm -rf server

// src: https://github.com/exogen/graphql-markdown
// install using the CLI command: `npm install -g graphql-markdown`
//go:generate sh -c "graphql-markdown schema.graphql > schema.md"
