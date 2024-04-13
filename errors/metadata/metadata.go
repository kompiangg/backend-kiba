package metadata

type Metadata struct {
	http
}

func (m *Metadata) SetHTTPMetadata(code int) *Metadata {
	m.http.setHTTPError(code)

	return m
}

func (m *Metadata) GetHTTPCode() int {
	return m.getHTTPCode()
}

func (m *Metadata) GetHTTPStatus() string {
	return m.getHTTPStatus()
}
