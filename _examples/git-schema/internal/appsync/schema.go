package appsync

//go:generate go run ../../../../. -schemas https://github.com/lkysow/sc.git/schema1.gql -out models.go -entities Entity1

type RawEntity struct {
	ID string
}

func CopyEntity(e Entity1) RawEntity {
	return RawEntity{ID: e.Id}
}
