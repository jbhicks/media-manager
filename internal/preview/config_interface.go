package preview

type ConfigProvider interface {
	GetThumbnailDir() string
	GetThumbnailSize() int
}
