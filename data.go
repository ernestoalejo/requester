package requester

type Data interface {
	Key() string
}

func GetData() Data {
	return nil
}

func SetData(data Data) {

}
