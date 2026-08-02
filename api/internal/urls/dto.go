package urls

type UrlShort struct {
	LongUrl string `json:"long_url" validate:"required,url"`
}
