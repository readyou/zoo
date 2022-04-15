package util

var Uploader *uploader = &uploader{}

type uploader struct {
}

func (*uploader) Upload(byteList []byte, base string) (string, error) {
	return "", nil
}
