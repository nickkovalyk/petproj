package mappers

//type MapperModel interface {
//	GetPKName() string
//}
//type MapperInterface interface {
//	GetTableName() string
//	Save(*MapperModel) (*MapperModel, error)
//	Delete(int) error
//}

type NotFoundError string

func (e NotFoundError) Error() string {
	return string(e)
}
