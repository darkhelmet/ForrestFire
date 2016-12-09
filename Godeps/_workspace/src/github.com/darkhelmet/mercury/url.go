package mercury

import (
	"encoding/json"
	"net/url"
)

type URL struct {
	url.URL
}

func (u *URL) UnmarshalJSON(data []byte) error {
	var rawUrl string
	if err := json.Unmarshal(data, &rawUrl); err != nil {
		return err
	}
	uri, err := url.Parse(rawUrl)
	u.URL = *uri
	return err
}

func (u *URL) MarshalJSON() ([]byte, error) {
	return []byte(u.String()), nil
}
