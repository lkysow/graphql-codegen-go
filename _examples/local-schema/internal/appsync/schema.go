package appsync

//go:generate go run ../../../../ -schemas ../../schema.graphql -entities Entity1 -out models.go

type RawEntity struct {
	ID string
}

func CopyEntity(e Entity1) RawEntity {
	return RawEntity{ID: e.Id}
}
