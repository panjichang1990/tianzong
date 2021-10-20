package tianzong

type IAdmin interface {
	ToJson() string
	CheckUri(uri string) bool
}
